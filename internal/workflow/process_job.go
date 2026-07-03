package workflow

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
	"github.com/0x4a5700/actions-metrics-converter/pkg/github"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func ProcessJob(payload github.WorkflowJobPayload) telemetry.JobSpans {
	job := payload.WorkflowJob
	return telemetry.JobSpans{
		TraceID:    traceIDFromRunID(job.RunId),
		ParentID:   spanIDFromInt(job.RunId),
		JobID:      spanIDFromInt(job.Id),
		Name:       job.Name,
		QueueStart: job.CreatedAt,
		QueueEnd:   job.StartedAt,
		RunStart:   job.StartedAt,
		RunEnd:     job.CompletedAt,
		Status:     job.Status,
		Conclusion: job.Conclusion,
		Attributes: []attribute.KeyValue{
			attribute.String("ci.job.conclusion", job.Conclusion),
			attribute.String("ci.job.labels", strings.Join(job.Labels, ",")),
			attribute.Int("ci.run.attempt", job.RunAttempt),
			attribute.String("ci.runner.name", job.RunnerName),
			attribute.String("ci.runner.group_name", job.RunnerGroupName),
			attribute.String("vcs.repository.full_name", payload.Repository.FullName),
			attribute.String("vcs.ref.head.name", job.HeadBranch),
			attribute.String("vcs.commit.sha", job.HeadSha),
		},
	}
}

func traceIDFromRunID(runID int) trace.TraceID {
	h := sha256.Sum256([]byte("run:" + strconv.Itoa(runID)))
	var id trace.TraceID
	copy(id[:], h[:16])
	return id
}

func spanIDFromInt(id int) trace.SpanID {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(id))
	return trace.SpanID(b)
}
