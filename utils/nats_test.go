package utils

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func TestExtractsNatsTraceContextFromValidHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := nats.Msg{
		Header: nats.Header{
			"traceparent": []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			"tracestate":  []string{"key1=value1,key2=value2"},
		},
	}

	resultCtx := GetNatsTraceContext(ctx, msg)

	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", carrier["traceparent"])
	assert.Equal(t, "key1=value1,key2=value2", carrier["tracestate"])
}

func TestHandlesNatsHeadersWithEmptyValuesGracefully(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := nats.Msg{
		Header: nats.Header{
			"traceparent": []string{""},
		},
	}

	resultCtx := GetNatsTraceContext(ctx, msg)
	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Empty(t, carrier["traceparent"])
}

func TestInjectsTraceContextIntoNatsHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	propagator := propagation.TraceContext{}
	inCarrier := propagation.MapCarrier{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"tracestate":  "key1=value1,key2=value2",
	}
	ctxWithTrace := propagator.Extract(ctx, inCarrier)

	headers := SetNatsHeaderTraceContext(ctxWithTrace)
	assert.Equal(t, inCarrier["traceparent"], headers.Get("traceparent"))
	assert.Equal(t, inCarrier["tracestate"], headers.Get("tracestate"))
}
