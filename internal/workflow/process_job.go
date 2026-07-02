package workflow

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"

	"github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
	"github.com/0x4a5700/actions-metrics-converter/pkg/github"
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
