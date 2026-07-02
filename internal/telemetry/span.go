package telemetry

import (
	"time"

	"go.opentelemetry.io/otel/trace"
)

type JobSpans struct {
	TraceID  trace.TraceID
	ParentID trace.SpanID
	JobID    trace.SpanID

	Name string

	QueueStart time.Time
	QueueEnd   time.Time
	RunStart   time.Time
	RunEnd     time.Time

	Status     string
	Conclusion string
}
