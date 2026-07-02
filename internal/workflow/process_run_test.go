package workflow

import (
	"testing"
	"time"

	"github.com/0x4a5700/actions-metrics-converter/pkg/github"
	"github.com/stretchr/testify/assert"
)

func TestProcessRun(t *testing.T) {
	queued := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	started := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	completed := time.Date(2026, 1, 1, 0, 5, 0, 0, time.UTC)

	tests := []struct {
		name       string
		run        github.WorkflowRun
		wantName   string
		wantStatus string
		wantConc   string
		wantQStart time.Time
		wantQEnd   time.Time
		wantRStart time.Time
		wantREnd   time.Time
	}{
		{
			name: "completed successful run",
			run: github.WorkflowRun{
				Id:           999,
				Name:         "build-and-push",
				Status:       "completed",
				Conclusion:   "success",
				CreatedAt:    queued,
				RunStartedAt: started,
				UpdatedAt:    completed,
			},
			wantName:   "build-and-push",
			wantStatus: "completed",
			wantConc:   "success",
			wantQStart: queued,
			wantQEnd:   started,
			wantRStart: started,
			wantREnd:   completed,
		},
		{
			name: "completed failed run",
			run: github.WorkflowRun{
				Id:           888,
				Name:         "deploy",
				Status:       "completed",
				Conclusion:   "failure",
				CreatedAt:    queued,
				RunStartedAt: started,
				UpdatedAt:    completed,
			},
			wantName:   "deploy",
			wantStatus: "completed",
			wantConc:   "failure",
			wantQStart: queued,
			wantQEnd:   started,
			wantRStart: started,
			wantREnd:   completed,
		},
		{
			name: "in progress run has zero end timestamps",
			run: github.WorkflowRun{
				Id:        777,
				Name:      "test",
				Status:    "in_progress",
				CreatedAt: queued,
			},
			wantName:   "test",
			wantStatus: "in_progress",
			wantQStart: queued,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessRun(github.WorkflowJobPayload{WorkflowRun: tt.run})

			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantConc, got.Conclusion)
			assert.True(t, got.QueueStart.Equal(tt.wantQStart), "QueueStart: got %v want %v", got.QueueStart, tt.wantQStart)
			assert.True(t, got.QueueEnd.Equal(tt.wantQEnd), "QueueEnd: got %v want %v", got.QueueEnd, tt.wantQEnd)
			assert.True(t, got.RunStart.Equal(tt.wantRStart), "RunStart: got %v want %v", got.RunStart, tt.wantRStart)
			assert.True(t, got.RunEnd.Equal(tt.wantREnd), "RunEnd: got %v want %v", got.RunEnd, tt.wantREnd)
		})
	}
}

func TestProcessRunTraceIDDeterminism(t *testing.T) {
	tests := []struct {
		name     string
		runIDA   int
		runIDB   int
		wantSame bool
	}{
		{"same run id produces same trace ID", 123, 123, true},
		{"different run ids produce different trace IDs", 123, 456, false},
		{"zero and non-zero differ", 0, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ProcessRun(github.WorkflowJobPayload{WorkflowRun: github.WorkflowRun{Id: tt.runIDA}}).TraceID
			b := ProcessRun(github.WorkflowJobPayload{WorkflowRun: github.WorkflowRun{Id: tt.runIDB}}).TraceID
			if tt.wantSame {
				assert.Equal(t, a, b)
			} else {
				assert.NotEqual(t, a, b)
			}
		})
	}
}

func TestProcessRunTraceIDMatchesJobTraceID(t *testing.T) {
	// Job spans and run spans for the same run_id must share a TraceID so
	// they are grouped together in the trace backend.
	runID := 28280744806

	jobSpans := ProcessJob(github.WorkflowJobPayload{
		WorkflowJob: github.WorkflowJob{RunId: runID},
	})
	runSpans := ProcessRun(github.WorkflowJobPayload{
		WorkflowRun: github.WorkflowRun{Id: runID},
	})

	assert.Equal(t, jobSpans.TraceID, runSpans.TraceID, "job and run spans must share a TraceID for the same run")
}

func TestProcessRunIDDerivation(t *testing.T) {
	tests := []struct {
		name  string
		runID int
	}{
		{"small id", 1},
		{"large id", 28280744806},
		{"zero", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessRun(github.WorkflowJobPayload{
				WorkflowRun: github.WorkflowRun{Id: tt.runID},
			})

			assert.Equal(t, traceIDFromRunID(tt.runID), got.TraceID)
			assert.Equal(t, spanIDFromInt(tt.runID), got.SpanID)
		})
	}
}
