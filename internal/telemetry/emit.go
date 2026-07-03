package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/codes"
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
		trace.WithAttributes(s.Attributes...),
	)
	setStatus(queueSpan, s.Conclusion)
	queueSpan.End(trace.WithTimestamp(s.QueueEnd))

	_, runSpan := tracer.Start(remoteCtx, s.Name+": run",
		trace.WithTimestamp(s.RunStart),
		trace.WithAttributes(s.Attributes...),
	)
	setStatus(runSpan, s.Conclusion)
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
		trace.WithAttributes(s.Attributes...),
	)
	setStatus(queueSpan, s.Conclusion)
	queueSpan.End(trace.WithTimestamp(s.QueueEnd))

	_, runSpan := tracer.Start(remoteCtx, s.Name+": run",
		trace.WithTimestamp(s.RunStart),
		trace.WithAttributes(s.Attributes...),
	)
	setStatus(runSpan, s.Conclusion)
	runSpan.End(trace.WithTimestamp(s.RunEnd))
}

func setStatus(span trace.Span, conclusion string) {
	switch conclusion {
	case "failure", "cancelled", "timed_out", "action_required":
		span.SetStatus(codes.Error, conclusion)
	case "success", "skipped", "neutral":
		span.SetStatus(codes.Ok, "")
	}
}
