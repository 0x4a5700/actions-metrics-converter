package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func newRecorder() (*tracetest.InMemoryExporter, trace.Tracer) {
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	return exp, tp.Tracer("test")
}

func TestEmitSpanCount(t *testing.T) {
	exp, tracer := newRecorder()
	Emit(context.Background(), tracer, JobSpans{
		TraceID:    trace.TraceID{1},
		ParentID:   trace.SpanID{1},
		Name:       "build",
		QueueStart: time.Now(),
		QueueEnd:   time.Now().Add(time.Minute),
		RunStart:   time.Now().Add(time.Minute),
		RunEnd:     time.Now().Add(5 * time.Minute),
	})

	assert.Len(t, exp.GetSpans(), 2)
}

func TestEmitSpanNames(t *testing.T) {
	tests := []struct {
		name      string
		jobName   string
		wantNames []string
	}{
		{
			name:      "standard job",
			jobName:   "build_docker",
			wantNames: []string{"build_docker: queue", "build_docker: run"},
		},
		{
			name:      "job with spaces",
			jobName:   "run tests",
			wantNames: []string{"run tests: queue", "run tests: run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tracer := newRecorder()
			Emit(context.Background(), tracer, JobSpans{
				TraceID:  trace.TraceID{1},
				ParentID: trace.SpanID{1},
				Name:     tt.jobName,
			})

			spans := exp.GetSpans()
			require.Len(t, spans, len(tt.wantNames))
			for i, want := range tt.wantNames {
				assert.Equal(t, want, spans[i].Name)
			}
		})
	}
}

func TestEmitSpanTimestamps(t *testing.T) {
	qStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	qEnd := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	rStart := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	rEnd := time.Date(2026, 1, 1, 0, 5, 0, 0, time.UTC)

	exp, tracer := newRecorder()
	Emit(context.Background(), tracer, JobSpans{
		TraceID:    trace.TraceID{1},
		ParentID:   trace.SpanID{1},
		Name:       "build",
		QueueStart: qStart,
		QueueEnd:   qEnd,
		RunStart:   rStart,
		RunEnd:     rEnd,
	})

	spans := exp.GetSpans()
	require.Len(t, spans, 2)

	tests := []struct {
		name      string
		span      tracetest.SpanStub
		wantStart time.Time
		wantEnd   time.Time
	}{
		{"queue span", spans[0], qStart, qEnd},
		{"run span", spans[1], rStart, rEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.span.StartTime.Equal(tt.wantStart), "StartTime: got %v want %v", tt.span.StartTime, tt.wantStart)
			assert.True(t, tt.span.EndTime.Equal(tt.wantEnd), "EndTime: got %v want %v", tt.span.EndTime, tt.wantEnd)
		})
	}
}

func TestEmitTraceAndParentIDs(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03}
	parentID := trace.SpanID{0x0a, 0x0b, 0x0c}

	exp, tracer := newRecorder()
	Emit(context.Background(), tracer, JobSpans{
		TraceID:  traceID,
		ParentID: parentID,
		Name:     "build",
	})

	spans := exp.GetSpans()
	require.Len(t, spans, 2)
	for _, span := range spans {
		assert.Equal(t, traceID, span.SpanContext.TraceID(), "span %q TraceID", span.Name)
		assert.Equal(t, parentID, span.Parent.SpanID(), "span %q ParentSpanID", span.Name)
	}
}

func TestEmitRunSpanCount(t *testing.T) {
	exp, tracer := newRecorder()
	EmitRun(context.Background(), tracer, RunSpans{
		TraceID: trace.TraceID{1},
		SpanID:  trace.SpanID{1},
		Name:    "build-and-push",
	})

	assert.Len(t, exp.GetSpans(), 2)
}

func TestEmitRunSpanNames(t *testing.T) {
	tests := []struct {
		name      string
		runName   string
		wantNames []string
	}{
		{
			name:      "standard workflow",
			runName:   "build-and-push",
			wantNames: []string{"build-and-push: queue", "build-and-push: run"},
		},
		{
			name:      "workflow with spaces",
			runName:   "deploy to prod",
			wantNames: []string{"deploy to prod: queue", "deploy to prod: run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp, tracer := newRecorder()
			EmitRun(context.Background(), tracer, RunSpans{
				TraceID: trace.TraceID{1},
				SpanID:  trace.SpanID{1},
				Name:    tt.runName,
			})

			spans := exp.GetSpans()
			require.Len(t, spans, len(tt.wantNames))
			for i, want := range tt.wantNames {
				assert.Equal(t, want, spans[i].Name)
			}
		})
	}
}

func TestEmitRunSpanTimestamps(t *testing.T) {
	qStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	qEnd := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	rStart := time.Date(2026, 1, 1, 0, 1, 0, 0, time.UTC)
	rEnd := time.Date(2026, 1, 1, 0, 5, 0, 0, time.UTC)

	exp, tracer := newRecorder()
	EmitRun(context.Background(), tracer, RunSpans{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{1},
		Name:       "build",
		QueueStart: qStart,
		QueueEnd:   qEnd,
		RunStart:   rStart,
		RunEnd:     rEnd,
	})

	spans := exp.GetSpans()
	require.Len(t, spans, 2)

	tests := []struct {
		name      string
		span      tracetest.SpanStub
		wantStart time.Time
		wantEnd   time.Time
	}{
		{"queue span", spans[0], qStart, qEnd},
		{"run span", spans[1], rStart, rEnd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.span.StartTime.Equal(tt.wantStart), "StartTime: got %v want %v", tt.span.StartTime, tt.wantStart)
			assert.True(t, tt.span.EndTime.Equal(tt.wantEnd), "EndTime: got %v want %v", tt.span.EndTime, tt.wantEnd)
		})
	}
}

func TestEmitSpanStatus(t *testing.T) {
	tests := []struct {
		conclusion string
		wantCode   codes.Code
	}{
		{"success", codes.Ok},
		{"skipped", codes.Ok},
		{"neutral", codes.Ok},
		{"failure", codes.Error},
		{"cancelled", codes.Error},
		{"timed_out", codes.Error},
		{"action_required", codes.Error},
		{"", codes.Unset},
	}

	for _, tt := range tests {
		t.Run(tt.conclusion, func(t *testing.T) {
			exp, tracer := newRecorder()
			Emit(context.Background(), tracer, JobSpans{
				TraceID:    trace.TraceID{1},
				ParentID:   trace.SpanID{1},
				Name:       "build",
				Conclusion: tt.conclusion,
			})

			spans := exp.GetSpans()
			require.Len(t, spans, 2)
			for _, span := range spans {
				assert.Equal(t, tt.wantCode, span.Status.Code, "span %q status", span.Name)
			}
		})
	}
}

func TestEmitSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		attribute.String("ci.job.conclusion", "success"),
		attribute.String("vcs.repository.full_name", "org/repo"),
	}

	exp, tracer := newRecorder()
	Emit(context.Background(), tracer, JobSpans{
		TraceID:    trace.TraceID{1},
		ParentID:   trace.SpanID{1},
		Name:       "build",
		Attributes: attrs,
	})

	spans := exp.GetSpans()
	require.Len(t, spans, 2)
	for _, span := range spans {
		assert.Subset(t, span.Attributes, attrs, "span %q missing attributes", span.Name)
	}
}

func TestEmitRunSpanStatus(t *testing.T) {
	tests := []struct {
		conclusion string
		wantCode   codes.Code
	}{
		{"success", codes.Ok},
		{"failure", codes.Error},
		{"cancelled", codes.Error},
		{"", codes.Unset},
	}

	for _, tt := range tests {
		t.Run(tt.conclusion, func(t *testing.T) {
			exp, tracer := newRecorder()
			EmitRun(context.Background(), tracer, RunSpans{
				TraceID:    trace.TraceID{1},
				SpanID:     trace.SpanID{1},
				Name:       "build",
				Conclusion: tt.conclusion,
			})

			spans := exp.GetSpans()
			require.Len(t, spans, 2)
			for _, span := range spans {
				assert.Equal(t, tt.wantCode, span.Status.Code, "span %q status", span.Name)
			}
		})
	}
}

func TestEmitRunSpanAttributes(t *testing.T) {
	attrs := []attribute.KeyValue{
		attribute.String("ci.run.conclusion", "failure"),
		attribute.String("vcs.repository.full_name", "org/repo"),
	}

	exp, tracer := newRecorder()
	EmitRun(context.Background(), tracer, RunSpans{
		TraceID:    trace.TraceID{1},
		SpanID:     trace.SpanID{1},
		Name:       "build",
		Attributes: attrs,
	})

	spans := exp.GetSpans()
	require.Len(t, spans, 2)
	for _, span := range spans {
		assert.Subset(t, span.Attributes, attrs, "span %q missing attributes", span.Name)
	}
}

func TestEmitRunTraceID(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03}
	spanID := trace.SpanID{0x0a, 0x0b, 0x0c}

	exp, tracer := newRecorder()
	EmitRun(context.Background(), tracer, RunSpans{
		TraceID: traceID,
		SpanID:  spanID,
		Name:    "build",
	})

	spans := exp.GetSpans()
	require.Len(t, spans, 2)
	for _, span := range spans {
		assert.Equal(t, traceID, span.SpanContext.TraceID(), "span %q TraceID", span.Name)
		assert.Equal(t, spanID, span.Parent.SpanID(), "span %q ParentSpanID", span.Name)
	}
}
