package otelemetry

import (
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Config holds the configuration for the telemetry setup.
type Config struct {
	// Service configuration.
	Service
	// Collector configuration.
	Collector
	// Flag to enable tracing.
	WithTraces bool
	// Flag to enable metrics.
	WithMetrics bool
	// Flag to enable logging.
	WithLogs bool
	// Options for resource configuration.
	ResourceOptions []sdkresource.Option
	// Options for tracer configuration.
	TracerOptions TracerOptions
	// Options for logger configuration.
	LoggerOptions LoggerOptions
	// Options for metric configuration.
	MetricOptions MetricOptions
}

// Service holds the service-related configuration.
type Service struct {
	// Name of the service.
	Name string
	// Namespace of the service.
	Namespace string
	// Version of the service.
	Version string
}

// Collector holds the collector-related configuration.
type Collector struct {
	Host string
	Port string
}

// LoggerOptions holds the options for logger configuration.
type LoggerOptions struct {
	// Options for the OTLP log exporter.
	ExporterOption []otlploggrpc.Option
	// Options for the logger provider.
	ProviderOption []sdklog.LoggerProviderOption
	// Options for the logger.
	LoggerOption []log.LoggerOption
}

// TracerOptions holds the options for tracer configuration.
type TracerOptions struct {
	// Options for the OTLP trace client.
	ClientOption []otlptracegrpc.Option
	// Options for the tracer provider.
	ProviderOption []sdktrace.TracerProviderOption
	// Options for the batch span processor.
	BatchSpanProcessorOption []sdktrace.BatchSpanProcessorOption
	// Options for the tracer.
	TracerOption []trace.TracerOption
}

// MetricOptions holds the options for metric configuration.
type MetricOptions struct {
	// Options for the OTLP metric exporter.
	ExporterOptions []otlpmetricgrpc.Option
	// Options for the metric provider.
	ProviderOptions []sdkmetric.Option
	// Options for the meter.
	MeterOptions []metric.MeterOption

	PeriodicInterval time.Duration
}
