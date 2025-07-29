package otelemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/resource"
)

func TestNewTelemetryWithMissingServiceName(t *testing.T) {
	cfg := Config{
		Service: Service{
			Name:      "",
			Namespace: "test-namespace",
			Version:   "1.0.0",
		},
		Collector: Collector{
			Host: "localhost",
			Port: "4317",
		},
	}

	_, err := New(cfg)
	assert.Error(t, err)
}

func TestNewTelemetryWithInvalidResourceOptions(t *testing.T) {
	cfg := Config{
		Service: Service{
			Name:      "test-service",
			Namespace: "test-namespace",
			Version:   "1.0.0",
		},
		Collector: Collector{
			Host: "localhost",
			Port: "4317",
		},
		ResourceOptions: []resource.Option{resource.WithAttributes()},
	}

	_, err := New(cfg)
	assert.Error(t, err)
}

func TestShutdownTelemetryWithTimeout(t *testing.T) {
	cfg := Config{
		Service: Service{
			Name:      "test-service",
			Namespace: "test-namespace",
			Version:   "1.0.0",
		},
		Collector: Collector{
			Host: "localhost",
			Port: "4317",
		},
	}

	tel, err := New(cfg)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err = tel.Shutdown(ctx)
	assert.Error(t, err)
}

func TestNewTelemetryWithResources(t *testing.T) {
	cfg := Config{
		Service: Service{
			Name:      "test-service",
			Namespace: "test-namespace",
			Version:   "1.0.0",
		},
		Collector: Collector{
			Host: "localhost",
			Port: "4317",
		},
		ResourceOptions: getResources("test-pod"),
	}

	_, err := New(cfg)
	assert.NoError(t, err)
}

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
