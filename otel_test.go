package otelemetry

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// newTestTelemetry creates a stdout-mode Telemetry instance for testing.
// It also sets the global OTel propagator, tracer, meter and logger providers.
func newTestTelemetry(t *testing.T) Telemetry {
	t.Helper()
	tel, err := New(Config{
		Service: Service{
			Name:      "test-service",
			Namespace: "test-namespace",
			Version:   "1.0.0",
		},
	})
	require.NoError(t, err)
	return tel
}

// --- New() ---

func TestNewTelemetryWithMissingServiceName(t *testing.T) {
	cfg := Config{
		Service: Service{
			Name:      "",
			Namespace: "test-namespace",
			Version:   "1.0.0",
		},
		Collector: Collector{Host: "localhost", Port: "4317"},
	}

	_, err := New(cfg)
	assert.Error(t, err)
}

func TestNewTelemetryWithEmptyResourceOptions(t *testing.T) {
	cfg := Config{
		Service:         Service{Name: "test-service", Namespace: "test-namespace", Version: "1.0.0"},
		Collector:       Collector{Host: "localhost", Port: "4317"},
		ResourceOptions: []resource.Option{resource.WithAttributes()},
	}

	_, err := New(cfg)
	assert.NoError(t, err)
}

func TestNewTelemetryWithResources(t *testing.T) {
	cfg := Config{
		Service:         Service{Name: "test-service", Namespace: "test-namespace", Version: "1.0.0"},
		Collector:       Collector{Host: "localhost", Port: "4317"},
		ResourceOptions: getResources("test-pod"),
	}

	_, err := New(cfg)
	assert.NoError(t, err)
}

func TestNewTelemetryStdoutMode(t *testing.T) {
	// No WithTraces/WithMetrics/WithLogs → falls back to stdout exporters
	tel, err := New(Config{
		Service: Service{Name: "svc", Namespace: "ns", Version: "0.1.0"},
	})
	require.NoError(t, err)
	assert.NotNil(t, tel)
	assert.NoError(t, tel.Shutdown(context.Background()))
}

// --- Shutdown() ---

func TestShutdownTelemetryWithTimeout(t *testing.T) {
	tel, err := New(Config{
		Service:   Service{Name: "test-service", Namespace: "test-namespace", Version: "1.0.0"},
		Collector: Collector{Host: "localhost", Port: "4317"},
	})
	require.NoError(t, err)

	err = tel.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestShutdownWithCancelledContext(t *testing.T) {
	tel, err := New(Config{
		Service:   Service{Name: "test-service", Namespace: "test-namespace", Version: "1.0.0"},
		Collector: Collector{Host: "localhost", Port: "4317"},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before Shutdown is called

	err = tel.Shutdown(ctx)
	assert.Error(t, err)
}

// --- Trace / Span ---

func TestStartSpanReturnsValidSpan(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	ctx, span := tel.Trace().StartSpan(context.Background(), "test-op")
	defer span.End()

	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
	assert.True(t, span.Span().SpanContext().IsValid())
}

func TestSpanTraceIDAndSpanID(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	_, span := tel.Trace().StartSpan(context.Background(), "test-op")
	defer span.End()

	traceID := span.TraceID()
	spanID := span.SpanID()

	assert.NotEmpty(t, traceID)
	assert.NotEmpty(t, spanID)
	assert.NotEqual(t, "00000000000000000000000000000000", traceID)
	assert.NotEqual(t, "0000000000000000", spanID)
}

func TestSpanFromContext(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	ctx, original := tel.Trace().StartSpan(context.Background(), "test-op")
	defer original.End()

	retrieved := tel.Trace().SpanFromContext(ctx)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.SpanID(), retrieved.SpanID())
}

func TestSpanAddEvent(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	_, span := tel.Trace().StartSpan(context.Background(), "test-op")
	defer span.End()

	assert.NotPanics(t, func() {
		span.AddEvent("something happened", Attribute("key", "value"))
	})
}

func TestSpanAddErrorEvent(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	_, span := tel.Trace().StartSpan(context.Background(), "test-op")
	defer span.End()

	assert.NotPanics(t, func() {
		span.AddErrorEvent("db failed", errors.New("connection refused"), Attribute("component", "db"))
	})
}

func TestSpanSetAttribute(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	_, span := tel.Trace().StartSpan(context.Background(), "test-op")
	defer span.End()

	assert.NotPanics(t, func() {
		span.SetAttribute(Attribute("user.id", 42), Attribute("user.name", "alice"))
	})
}

func TestSpanRecordError(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	_, span := tel.Trace().StartSpan(context.Background(), "test-op")
	defer span.End()

	assert.NotPanics(t, func() {
		span.RecordError(errors.New("unexpected failure"))
	})
}

func TestContextWithSpan(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	_, original := tel.Trace().StartSpan(context.Background(), "original")
	defer original.End()

	ctx2 := tel.Trace().ContextWithSpan(context.Background(), original.Span())
	retrieved := tel.Trace().SpanFromContext(ctx2)
	require.NotNil(t, retrieved)
	assert.Equal(t, original.SpanID(), retrieved.SpanID())
}

func TestContextWithRemoteSpanContext(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	_, span := tel.Trace().StartSpan(context.Background(), "original")
	defer span.End()

	ctx2 := tel.Trace().ContextWithRemoteSpanContext(context.Background(), span.Span())
	assert.NotNil(t, ctx2)
}

// --- Attribute() ---

type stringerVal struct{ v string }

func (s stringerVal) String() string { return s.v }

func TestAttribute(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		wantType attribute.Type
	}{
		{"string", "hello", attribute.STRING},
		{"int", 42, attribute.INT64},
		{"int64", int64(42), attribute.INT64},
		{"bool", true, attribute.BOOL},
		{"float64", 3.14, attribute.FLOAT64},
		{"string_slice", []string{"a", "b"}, attribute.STRINGSLICE},
		{"int_slice", []int{1, 2}, attribute.INT64SLICE},
		{"int64_slice", []int64{1, 2}, attribute.INT64SLICE},
		{"bool_slice", []bool{true, false}, attribute.BOOLSLICE},
		{"float64_slice", []float64{1.1, 2.2}, attribute.FLOAT64SLICE},
		{"stringer", stringerVal{"hi"}, attribute.STRING},
		{"unknown_type", struct{ X int }{1}, attribute.STRING}, // fallback: fmt.Sprintf
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := Attribute("k", tt.value)
			assert.Equal(t, attribute.Key("k"), attr.Key)
			assert.Equal(t, tt.wantType, attr.Value.Type())
		})
	}
}

// --- LogAttribute() ---

func TestLogAttribute(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		wantKind log.Kind
	}{
		{"string", "hello", log.KindString},
		{"int", 42, log.KindInt64},
		{"int64", int64(42), log.KindInt64},
		{"bool", true, log.KindBool},
		{"float64", 3.14, log.KindFloat64},
		{"unknown_type", struct{}{}, log.KindString}, // fallback
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr := LogAttribute("k", tt.value)
			assert.Equal(t, "k", attr.Key)
			assert.Equal(t, tt.wantKind, attr.Value.Kind())
		})
	}
}

// --- Propagation ---

func TestInjectExtract(t *testing.T) {
	tel := newTestTelemetry(t) // sets global propagator
	defer tel.Shutdown(context.Background())

	ctx, span := tel.Trace().StartSpan(context.Background(), "test-op")
	defer span.End()

	headers := make(map[string]string)
	Inject(ctx, headers)
	assert.NotEmpty(t, headers["traceparent"])

	newCtx := Extract(context.Background(), headers)
	extracted := tel.Trace().SpanFromContext(newCtx)
	require.NotNil(t, extracted)
	assert.True(t, extracted.Span().SpanContext().IsValid())
}

func TestInjectExtractHTTPHeaders(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	ctx, span := tel.Trace().StartSpan(context.Background(), "http-op")
	defer span.End()

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	InjectHTTPHeaders(ctx, req.Header)
	assert.NotEmpty(t, req.Header.Get("traceparent"))

	newCtx := ExtractHTTPHeaders(context.Background(), req.Header)
	extracted := tel.Trace().SpanFromContext(newCtx)
	require.NotNil(t, extracted)
	assert.True(t, extracted.Span().SpanContext().IsValid())
}

// --- Baggage ---

func TestBaggageAddBaggageItem(t *testing.T) {
	ctx := AddBaggageItem(context.Background(), "tenant", "acme")
	assert.Equal(t, "acme", GetBaggageItem(ctx, "tenant"))
}

func TestBaggageAddBaggageItems(t *testing.T) {
	ctx := AddBaggageItems(context.Background(), map[string]string{
		"env":    "prod",
		"region": "eu-west",
	})
	assert.Equal(t, "prod", GetBaggageItem(ctx, "env"))
	assert.Equal(t, "eu-west", GetBaggageItem(ctx, "region"))
}

func TestBaggageGetBaggage(t *testing.T) {
	ctx := AddBaggageItem(context.Background(), "key", "val")
	b := GetBaggage(ctx)
	assert.Equal(t, "val", b.Member("key").Value())
}

func TestBaggageRemoveBaggageItem(t *testing.T) {
	ctx := AddBaggageItems(context.Background(), map[string]string{"a": "1", "b": "2"})
	ctx = RemoveBaggageItem(ctx, "a")
	assert.Empty(t, GetBaggageItem(ctx, "a"))
	assert.Equal(t, "2", GetBaggageItem(ctx, "b"))
}

// --- Metrics ---

func TestMetricInt64Counter(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	counter, err := tel.Metric().Int64Counter("test_int64_counter")
	require.NoError(t, err)
	assert.NotPanics(t, func() {
		counter.Add(context.Background(), 1)
	})
}

func TestMetricFloat64Counter(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	counter, err := tel.Metric().Float64Counter("test_f64_counter")
	require.NoError(t, err)
	assert.NotPanics(t, func() {
		counter.Add(context.Background(), 3.14)
	})
}

func TestMetricFloat64Histogram(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	hist, err := tel.Metric().Float64Histogram("test_histogram")
	require.NoError(t, err)
	assert.NotPanics(t, func() {
		hist.Record(context.Background(), 0.5)
	})
}

func TestMetricInt64UpDownCounter(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())

	counter, err := tel.Metric().Int64UpDownCounter("test_updown")
	require.NoError(t, err)
	assert.NotPanics(t, func() {
		counter.Add(context.Background(), -1)
	})
}

// --- Logging ---

func TestLogAllLevels(t *testing.T) {
	tel := newTestTelemetry(t)
	defer tel.Shutdown(context.Background())
	ctx := context.Background()

	assert.NotPanics(t, func() {
		tel.Log().Debug(ctx, "debug message", LogAttribute("k", "v"))
		tel.Log().Info(ctx, "info message", LogAttribute("count", 42))
		tel.Log().Warning(ctx, "warning message")
		tel.Log().Error(ctx, "error message", LogAttribute("err", "boom"))
		tel.Log().Fatal(ctx, "fatal message")
	})
}

// --- helpers ---

func getResources(podname string) []resource.Option {
	if podname == "" {
		podname = "default-pod"
	}

	return []resource.Option{
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithAttributes(Attribute("pod.name", podname)),
	}
}
