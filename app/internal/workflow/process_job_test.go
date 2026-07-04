package workflow

import (
	"crypto/sha256"
	"strconv"
	"testing"
	"time"

	"github.com/0x4a5700/actions-metrics-converter/pkg/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestProcessJob(t *testing.T) {
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 1, 0, 5, 0, 0, time.UTC)

	tests := []struct {
		name       string
		job        github.WorkflowJob
		wantName   string
		wantStatus string
		wantConc   string
		wantQStart time.Time
		wantQEnd   time.Time
		wantRStart time.Time
		wantREnd   time.Time
	}{
		{
			name: "maps all fields",
			job: github.WorkflowJob{
				Id:          100,
				RunId:       999,
				Name:        "build_docker",
				Status:      "completed",
				Conclusion:  "success",
				CreatedAt:   t0,
				StartedAt:   t1,
				CompletedAt: t2,
			},
			wantName:   "build_docker",
			wantStatus: "completed",
			wantConc:   "success",
			wantQStart: t0,
			wantQEnd:   t1,
			wantRStart: t1,
			wantREnd:   t2,
		},
		{
			name: "failure conclusion",
			job: github.WorkflowJob{
				Id:          200,
				RunId:       888,
				Name:        "test",
				Status:      "completed",
				Conclusion:  "failure",
				CreatedAt:   t0,
				StartedAt:   t1,
				CompletedAt: t2,
			},
			wantName:   "test",
			wantStatus: "completed",
			wantConc:   "failure",
			wantQStart: t0,
			wantQEnd:   t1,
			wantRStart: t1,
			wantREnd:   t2,
		},
		{
			name: "queued job has zero run timestamps",
			job: github.WorkflowJob{
				Id:     300,
				RunId:  777,
				Name:   "lint",
				Status: "queued",
			},
			wantName:   "lint",
			wantStatus: "queued",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessJob(github.WorkflowJobPayload{WorkflowJob: tt.job})

			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantConc, got.Conclusion)
			assert.True(t, got.QueueStart.Equal(tt.wantQStart), "QueueStart")
			assert.True(t, got.QueueEnd.Equal(tt.wantQEnd), "QueueEnd")
			assert.True(t, got.RunStart.Equal(tt.wantRStart), "RunStart")
			assert.True(t, got.RunEnd.Equal(tt.wantREnd), "RunEnd")
		})
	}
}

func TestProcessJobTraceIDDeterminism(t *testing.T) {
	tests := []struct {
		name     string
		runIDA   int
		runIDB   int
		wantSame bool
	}{
		{"same run_id produces same trace ID", 123, 123, true},
		{"different run_ids produce different trace IDs", 123, 456, false},
		{"zero and non-zero differ", 0, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ProcessJob(github.WorkflowJobPayload{WorkflowJob: github.WorkflowJob{RunId: tt.runIDA}}).TraceID
			b := ProcessJob(github.WorkflowJobPayload{WorkflowJob: github.WorkflowJob{RunId: tt.runIDB}}).TraceID
			if tt.wantSame {
				assert.Equal(t, a, b)
			} else {
				assert.NotEqual(t, a, b)
			}
		})
	}
}

func TestProcessJobIDDerivation(t *testing.T) {
	tests := []struct {
		name  string
		jobID int
		runID int
	}{
		{"small ids", 1, 1},
		{"large ids", 28280744806, 183451495},
		{"zero job id", 0, 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessJob(github.WorkflowJobPayload{
				WorkflowJob: github.WorkflowJob{Id: tt.jobID, RunId: tt.runID},
			})

			assert.Equal(t, traceIDFromRunID(tt.runID), got.TraceID)
			assert.Equal(t, spanIDFromInt(tt.runID), got.ParentID)
			assert.Equal(t, spanIDFromInt(tt.jobID), got.JobID)
		})
	}
}

func TestProcessJobAttributes(t *testing.T) {
	tests := []struct {
		name     string
		payload  github.WorkflowJobPayload
		wantAttr map[string]string
	}{
		{
			name: "populates key attributes",
			payload: github.WorkflowJobPayload{
				WorkflowJob: github.WorkflowJob{
					Conclusion:      "success",
					Labels:          []string{"self-hosted", "linux"},
					RunAttempt:      2,
					RunnerName:      "runner-01",
					RunnerGroupName: "default",
					HeadBranch:      "main",
					HeadSha:         "abc123",
				},
				Repository: github.Repository{FullName: "org/repo"},
			},
			wantAttr: map[string]string{
				"ci.job.conclusion":        "success",
				"ci.job.labels":            "self-hosted,linux",
				"ci.runner.name":           "runner-01",
				"ci.runner.group_name":     "default",
				"vcs.repository.full_name": "org/repo",
				"vcs.ref.head.name":        "main",
				"vcs.commit.sha":           "abc123",
			},
		},
		{
			name: "empty labels produces empty string",
			payload: github.WorkflowJobPayload{
				WorkflowJob: github.WorkflowJob{Labels: []string{}},
			},
			wantAttr: map[string]string{
				"ci.job.labels": "",
			},
		},
		{
			name: "failure conclusion is recorded",
			payload: github.WorkflowJobPayload{
				WorkflowJob: github.WorkflowJob{Conclusion: "failure"},
			},
			wantAttr: map[string]string{
				"ci.job.conclusion": "failure",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessJob(tt.payload)
			for key, want := range tt.wantAttr {
				val, ok := findAttr(got.Attributes, key)
				require.True(t, ok, "attribute %q not found", key)
				assert.Equal(t, want, val.AsString(), "attribute %q", key)
			}
		})
	}
}

func findAttr(attrs []attribute.KeyValue, key string) (attribute.Value, bool) {
	for _, kv := range attrs {
		if string(kv.Key) == key {
			return kv.Value, true
		}
	}
	return attribute.Value{}, false
}

func TestTraceIDFromRunID(t *testing.T) {
	tests := []struct {
		name  string
		runID int
	}{
		{"zero", 0},
		{"typical", 28280744806},
		{"small", 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := traceIDFromRunID(tt.runID)

			h := sha256.Sum256([]byte("run:" + strconv.Itoa(tt.runID)))
			var want trace.TraceID
			copy(want[:], h[:16])

			assert.Equal(t, want, got)
		})
	}
}
