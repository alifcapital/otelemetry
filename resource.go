package otelemetry

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func newResource(ctx context.Context, cfg Config) (*sdkresource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.Service.Name),
		semconv.ServiceNamespaceKey.String(cfg.Service.Namespace),
		semconv.ServiceVersionKey.String(cfg.Service.Version),
	}
	return resource.New(ctx, resourceOpts(cfg.ResourceOptions, attrs)...)
}

func resourceOpts(options []sdkresource.Option, attrs []attribute.KeyValue) []sdkresource.Option {
	opts := []sdkresource.Option{
		sdkresource.WithAttributes(attrs...),
	}

	if len(options) > 0 {
		opts = append(opts, options...)
	}

	return opts
}
