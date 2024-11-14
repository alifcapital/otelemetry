package otelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Trace interface provides methods for tracing operations.
type Trace interface {
	// Trace returns the original tracer.
	Trace() trace.Tracer // original trace

	// StartSpan starts a new span with the given name and options.
	StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, Span)
	// SpanFromContext retrieves the span from the given context.
	SpanFromContext(ctx context.Context) Span

	// ContextWithSpan returns a new context with the given span.
	ContextWithSpan(ctx context.Context, span trace.Span) context.Context
	// ContextWithRemoteSpanContext returns a new context with the remote span context.
	ContextWithRemoteSpanContext(ctx context.Context, span trace.Span) context.Context
}

// oteltrace is an implementation of the Trace interface using OpenTelemetry.
type oteltrace struct {
	trace trace.Tracer
}

func (t *oteltrace) Trace() trace.Tracer {
	return t.trace
}

func (t *oteltrace) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, Span) {
	ctx, span := t.trace.Start(ctx, name, opts...)
	return ctx, &otelspan{span: span}
}

func (t *oteltrace) SpanFromContext(ctx context.Context) Span {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return nil
	}
	return &otelspan{span: span}
}

func (t *oteltrace) ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

func (t *oteltrace) ContextWithRemoteSpanContext(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithRemoteSpanContext(ctx, span.SpanContext())
}

func newTraceProvider(ctx context.Context, otelAgentAddr string, res *sdkresource.Resource, opts TracerOptions) (*sdktrace.TracerProvider, error) {
	client := otlptracegrpc.NewClient(traceClientOpts(otelAgentAddr, opts.ClientOption...)...)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter, opts.BatchSpanProcessorOption...)
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	return provider, nil
}

func newStdoutTraceProvider(res *sdkresource.Resource) (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New( /*stdouttrace.WithPrettyPrint()*/ )
	if err != nil {
		return nil, fmt.Errorf("creating stdout exporter: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	return tracerProvider, nil
}

func traceClientOpts(otelAgentAddr string, opts ...otlptracegrpc.Option) []otlptracegrpc.Option {
	options := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelAgentAddr),
	}

	if len(opts) > 0 {
		options = append(options, opts...)
	}

	return opts
}
