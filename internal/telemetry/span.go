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

type RunSpans struct {
	TraceID trace.TraceID
	SpanID  trace.SpanID

	Name string

	QueueStart time.Time // CreatedAt: when the run was queued
	QueueEnd   time.Time // RunStartedAt: when execution began
	RunStart   time.Time // RunStartedAt
	RunEnd     time.Time // UpdatedAt: when completed

	Status     string
	Conclusion string
}
