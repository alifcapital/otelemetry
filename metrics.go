package otelemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

// Metric interface provides methods for creating and managing various types of metrics.
type Metric interface {
	// Metric returns the underlying OpenTelemetry meter.
	Metric() metric.Meter

	Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error)
	Int64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error)
	Int64Histogram(name string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error)
	Int64Gauge(name string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error)
	Int64ObservableCounter(name string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error)
	Int64ObservableUpDownCounter(name string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error)
	Int64ObservableGauge(name string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error)
	Float64Counter(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error)
	Float64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error)
	Float64Histogram(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error)
	Float64Gauge(name string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error)
	Float64ObservableCounter(name string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error)
	Float64ObservableUpDownCounter(name string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error)
	Float64ObservableGauge(name string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error)
	RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error)
}

// otelmetric is an implementation of the Metric interface using OpenTelemetry.
type otelmetric struct {
	metric metric.Meter
}

func (m *otelmetric) Metric() metric.Meter {
	return m.metric
}

func (m *otelmetric) Int64Counter(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error) {
	return m.metric.Int64Counter(name, options...)
}

func (m *otelmetric) Int64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	return m.metric.Int64UpDownCounter(name, options...)
}

func (m *otelmetric) Int64Histogram(name string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error) {
	return m.metric.Int64Histogram(name, options...)
}

func (m *otelmetric) Int64Gauge(name string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	return m.metric.Int64Gauge(name, options...)
}

func (m *otelmetric) Int64ObservableCounter(name string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	return m.metric.Int64ObservableCounter(name, options...)
}

func (m *otelmetric) Int64ObservableUpDownCounter(name string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	return m.metric.Int64ObservableUpDownCounter(name, options...)
}

func (m *otelmetric) Int64ObservableGauge(name string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	return m.metric.Int64ObservableGauge(name, options...)
}

func (m *otelmetric) Float64Counter(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	return m.metric.Float64Counter(name, options...)
}

func (m *otelmetric) Float64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	return m.metric.Float64UpDownCounter(name, options...)
}

func (m *otelmetric) Float64Histogram(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
	return m.metric.Float64Histogram(name, options...)
}

func (m *otelmetric) Float64Gauge(name string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	return m.metric.Float64Gauge(name, options...)
}

func (m *otelmetric) Float64ObservableCounter(name string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	return m.metric.Float64ObservableCounter(name, options...)
}

func (m *otelmetric) Float64ObservableUpDownCounter(name string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	return m.metric.Float64ObservableUpDownCounter(name, options...)
}

func (m *otelmetric) Float64ObservableGauge(name string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	return m.metric.Float64ObservableGauge(name, options...)
}

func (m *otelmetric) RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error) {
	return m.metric.RegisterCallback(f, instruments...)
}

func newMeterProvider(ctx context.Context, otelAgentAddr string, res *sdkresource.Resource, opts MetricOptions) (*sdkmetric.MeterProvider, error) {
	exporter, err := otlpmetricgrpc.New(ctx, meterExporterOpts(otelAgentAddr, opts.ExporterOptions...)...)
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(meterProviderOpts(exporter, func() time.Duration {
		if opts.PeriodicInterval == 0 {
			return 5 * time.Second
		}
		// Use the default if PeriodicInterval is not set
		// This allows for a custom interval to be set in MetricOptions
		return opts.PeriodicInterval
	}(), res, opts.ProviderOptions...)...)

	return provider, nil
}

func newStdoutMeterProvider(res *sdkresource.Resource) (*sdkmetric.MeterProvider, error) {
	exporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
	)

	return provider, nil
}

func meterExporterOpts(otelAgentAddr string, opts ...otlpmetricgrpc.Option) []otlpmetricgrpc.Option {
	options := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint(otelAgentAddr),
	}

	if len(opts) == 0 {
		return options
	}

	return opts
}

func meterProviderOpts(exporter sdkmetric.Exporter, interval time.Duration, res *sdkresource.Resource, opts ...sdkmetric.Option) []sdkmetric.Option {
	options := []sdkmetric.Option{
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				exporter,
				sdkmetric.WithInterval(interval), //2*time.Second
			),
		),
	}

	if len(opts) > 0 {
		options = append(options, opts...)
	}

	return options
}
