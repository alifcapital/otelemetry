package otelemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Span represents an OpenTelemetry span and provides methods to interact with it.
type Span interface {
	// Span returns the underlying OpenTelemetry span.
	Span() trace.Span // think about naming

	// AddEvent adds an event to the span with the given name and attributes.
	AddEvent(name string, kv ...attribute.KeyValue)

	// AddErrorEvent adds an error event to the span with the given name, error, and attributes.
	AddErrorEvent(name string, err error, kv ...attribute.KeyValue)

	// SetAttribute sets an attribute on the span.
	SetAttribute(kv ...attribute.KeyValue)

	// End ends the span with the given options.
	End(opts ...trace.SpanEndOption) // aka Finish()

	// RecordError records an error on the span with the given attributes.
	RecordError(err error, kv ...attribute.KeyValue)

	// TraceID returns the trace ID of the span.
	TraceID() string

	// SpanID returns the span ID of the span.
	SpanID() string

	// Inject injects the span context into the given carrier.
	Inject(ctx context.Context, carrier propagation.TextMapCarrier)

	// Extract extracts the span context from the given carrier.
	Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context
}

// otelspan is an implementation of the Span interface using OpenTelemetry.
type otelspan struct {
	span trace.Span
}

func (s *otelspan) Span() trace.Span {
	return s.span
}

func (s *otelspan) AddEvent(name string, kv ...attribute.KeyValue) {
	s.span.AddEvent(name, trace.WithAttributes(kv...))
}

func (s *otelspan) AddErrorEvent(name string, err error, kv ...attribute.KeyValue) {
	s.span.SetStatus(codes.Error, err.Error())
	kv = append(kv, attribute.String("error.message", err.Error()))
	kv = append(kv, attribute.String("error.type", fmt.Sprintf("%T", err)))
	s.span.AddEvent(name, trace.WithAttributes(kv...))
}

func (s *otelspan) SetAttribute(kv ...attribute.KeyValue) {
	s.span.SetAttributes(kv...)
}

func (s *otelspan) RecordError(err error, kv ...attribute.KeyValue) {
	s.span.SetStatus(codes.Error, err.Error())
	s.span.RecordError(err, trace.WithAttributes(kv...))
}

func (s *otelspan) End(opts ...trace.SpanEndOption) {
	s.span.End(opts...)
}

func (s *otelspan) TraceID() string {
	return s.span.SpanContext().TraceID().String()
}

func (s *otelspan) SpanID() string {
	return s.span.SpanContext().SpanID().String()
}

// Deprecated: use Inject
func (s *otelspan) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, carrier)
}

// Deprecated: use Extract
func (s *otelspan) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, carrier)
}
