package workflow

import (
	"github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
	"github.com/0x4a5700/actions-metrics-converter/pkg/github"
	"go.opentelemetry.io/otel/attribute"
)

func ProcessRun(payload github.WorkflowJobPayload) telemetry.RunSpans {
	run := payload.WorkflowRun
	return telemetry.RunSpans{
		TraceID:    traceIDFromRunID(run.Id),
		SpanID:     spanIDFromInt(run.Id),
		Name:       run.Name,
		QueueStart: run.CreatedAt,
		QueueEnd:   run.RunStartedAt,
		RunStart:   run.RunStartedAt,
		RunEnd:     run.UpdatedAt,
		Status:     run.Status,
		Conclusion: run.Conclusion,
		Attributes: []attribute.KeyValue{
			attribute.String("ci.run.conclusion", run.Conclusion),
			attribute.Int("ci.run.id", run.Id),
			attribute.Int("ci.run.number", run.RunNumber),
			attribute.Int("ci.run.attempt", run.RunAttempt),
			attribute.String("ci.run.event", run.Event),
			attribute.String("vcs.repository.full_name", payload.Repository.FullName),
			attribute.String("vcs.ref.head.name", run.HeadBranch),
			attribute.String("vcs.commit.sha", run.HeadSha),
		},
	}
}
