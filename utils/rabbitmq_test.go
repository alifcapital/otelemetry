package utils

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func TestExtractsTraceContextFromHeadersWithValidStringValues(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := amqp.Delivery{
		Headers: amqp.Table{
			"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
	}

	resultCtx := GetRabbitMQTraceContext(ctx, msg)

	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", carrier["traceparent"])
}

func TestReturnsOriginalContextWhenHeadersAreAbsent(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := amqp.Delivery{
		Headers: amqp.Table{},
	}

	resultCtx := GetRabbitMQTraceContext(ctx, msg)
	assert.Equal(t, ctx, resultCtx)
}

func TestHandlesHeadersWithNonStringValuesGracefully(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := amqp.Delivery{
		Headers: amqp.Table{
			"traceparent": 12345, // Non-string value
		},
	}

	resultCtx := GetRabbitMQTraceContext(ctx, msg)
	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Empty(t, carrier["traceparent"])
}

func TestExtractsTraceContextFromHeadersWithMultipleValidKeys(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := amqp.Delivery{
		Headers: amqp.Table{
			"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			"tracestate":  "key1=value1,key2=value2",
		},
	}

	resultCtx := GetRabbitMQTraceContext(ctx, msg)

	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", carrier["traceparent"])
	assert.Equal(t, "key1=value1,key2=value2", carrier["tracestate"])
}

func TestHandlesHeadersWithEmptyStringValuesGracefully(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := amqp.Delivery{
		Headers: amqp.Table{
			"traceparent": "",
		},
	}

	resultCtx := GetRabbitMQTraceContext(ctx, msg)
	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Empty(t, carrier["traceparent"])
}

func TestInjectsTraceContextIntoHeadersWithMultipleKeys(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	// Simulate incoming headers
	msg := amqp.Delivery{
		Headers: amqp.Table{
			"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			"tracestate":  "key1=value1,key2=value2",
		},
	}
	ctx := context.Background()
	// Extract context from headers
	ctx = GetRabbitMQTraceContext(ctx, msg)
	// Inject context into new headers
	headers := SetRabbitMQHeaderTraceContext(ctx)

	assert.Equal(t, msg.Headers["traceparent"], headers["traceparent"])
	assert.Equal(t, msg.Headers["tracestate"], headers["tracestate"])
}
