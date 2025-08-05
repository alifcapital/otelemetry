package utils

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func GetRabbitMQTraceContext(ctx context.Context, msg amqp.Delivery) context.Context {
	carrier := propagation.MapCarrier{}
	for k, v := range msg.Headers {
		if str, ok := v.(string); ok {
			carrier[k] = str
		}
	}

	propagator := otel.GetTextMapPropagator()
	return propagator.Extract(ctx, carrier)
}

func SetRabbitMQHeaderTraceContext(ctx context.Context) amqp.Table {
	carrier := propagation.MapCarrier{}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, carrier)

	headers := amqp.Table{}
	for k, v := range carrier {
		headers[k] = v
	}

	return headers
}
