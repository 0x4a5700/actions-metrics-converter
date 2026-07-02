package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func Emit(ctx context.Context, tracer trace.Tracer, s JobSpans) {
	remoteCtx := trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    s.TraceID,
		SpanID:     s.ParentID,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	}))

	_, queueSpan := tracer.Start(remoteCtx, s.Name+": queue",
		trace.WithTimestamp(s.QueueStart),
	)
	queueSpan.End(trace.WithTimestamp(s.QueueEnd))

	_, runSpan := tracer.Start(remoteCtx, s.Name+": run",
		trace.WithTimestamp(s.RunStart),
	)
	runSpan.End(trace.WithTimestamp(s.RunEnd))
}
