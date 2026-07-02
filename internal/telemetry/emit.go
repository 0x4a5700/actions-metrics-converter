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

func EmitRun(ctx context.Context, tracer trace.Tracer, s RunSpans) {
	// s.SpanID is used as the phantom parent so the SDK assigns the right TraceID.
	// Job spans reference this same SpanID as their ParentID, so all spans share
	// a TraceID and a common parent reference. A custom IDGenerator on the
	// TracerProvider would be needed to give this span exactly SpanID=s.SpanID
	// and make job spans true children rather than siblings.
	remoteCtx := trace.ContextWithRemoteSpanContext(ctx, trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    s.TraceID,
		SpanID:     s.SpanID,
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
