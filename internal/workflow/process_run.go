package workflow

import (
	"github.com/0x4a5700/actions-metrics-converter/internal/telemetry"
	"github.com/0x4a5700/actions-metrics-converter/pkg/github"
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
	}
}
