package utils

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/alifcapital/otelemetry"
)

func GetJetstreamTraceContext(ctx context.Context, msg jetstream.Msg) context.Context {
	var headers = make(map[string]string)
	for k, _ := range msg.Headers() {
		headers[k] = msg.Headers().Get(k)
	}

	return otelemetry.Extract(ctx, headers)
}

func GetNatsTraceContext(ctx context.Context, msg nats.Msg) context.Context {
	var headers = make(map[string]string)
	for k, _ := range msg.Header {
		headers[k] = msg.Header.Get(k)
	}

	return otelemetry.Extract(ctx, headers)
}

func SetNatsHeaderTraceContext(ctx context.Context) nats.Header {
	carrier := propagation.MapCarrier{}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, carrier)

	var header = make(nats.Header)
	for key, value := range carrier {
		header.Add(key, value)
	}

	return header
}
