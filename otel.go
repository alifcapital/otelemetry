package otelemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Telemetry implements the OpenTelemetry API.
// t.Trace().StartSpan(ctx, "name")
// t.Log().Info(ctx, "message")
// t.Metric().NewInt64Counter("name")
type Telemetry interface {
	Trace() Trace
	Log() Log
	Metric() Metric
	Shutdown(ctx context.Context) error
}

type telemetry struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	tracer         trace.Tracer
	meter          metric.Meter
	logger         log.Logger
	serviceName    string
}

func (t *telemetry) Trace() Trace {
	return &oteltrace{trace: t.tracer}
}

func (t *telemetry) Log() Log {
	return &otellog{log: t.logger}
}

func (t *telemetry) Metric() Metric {
	return &otelmetric{metric: t.meter}
}

func (t *telemetry) Shutdown(ctx context.Context) error {
	cxt, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// pushes any last exports to the receiver
	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}

	if t.meterProvider != nil {
		if err := t.meterProvider.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}

	if t.loggerProvider != nil {
		if err := t.loggerProvider.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}

	return nil
}

func New(cfg Config) (Telemetry, error) {

	var (
		ctx               = context.Background()
		serviceName       = cfg.Service.Name
		otelCollectorHost = cfg.Collector.Host
		otelCollectorPort = cfg.Collector.Port

		tracerProvider *sdktrace.TracerProvider
		meterProvider  *sdkmetric.MeterProvider
		loggerProvider *sdklog.LoggerProvider

		err        error
		otelemetry = telemetry{
			serviceName: serviceName,
		}
	)

	// otel collector OTEL_COLLECTOR_HOST:OTEL_COLLECTOR_PORT_GRPC
	otelAgentAddr := fmt.Sprintf("%s:%s", otelCollectorHost, otelCollectorPort)

	// resource
	res, err := resource.New(ctx, resourceOpts(cfg, cfg.ResourceOptions)...)
	handleErr(err, "failed to create resource")

	// traces
	if cfg.WithTraces {
		tracerProvider, err = newTraceProvider(ctx, otelAgentAddr, res, cfg.TracerOptions)
		handleErr(err, "failed to create the collector trace exporter or provider")
	} else {
		tracerProvider, err = newStdoutTraceProvider(res)
		handleErr(err, "failed to create the collector trace exporter or provider")
	}

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.SetTracerProvider(tracerProvider)
	otelemetry.tracerProvider = tracerProvider
	otelemetry.tracer = tracerProvider.Tracer(serviceName, cfg.TracerOptions.TracerOption...)

	// metrics
	if cfg.WithMetrics {
		meterProvider, err = newMeterProvider(ctx, otelAgentAddr, res, cfg.MetricOptions)
		handleErr(err, "failed to create the collector metric exporter or provider - grpc")
	} else {
		meterProvider, err = newStdoutMeterProvider(res)
		handleErr(err, "failed to create the collector metric exporter or provider - stdout")
	}

	otel.SetMeterProvider(meterProvider)
	otelemetry.meterProvider = meterProvider
	otelemetry.meter = meterProvider.Meter(serviceName, cfg.MetricOptions.MeterOptions...)

	// logs - stdout or otlp
	if cfg.WithLogs {
		loggerProvider, err = newLoggerProvider(ctx, otelAgentAddr, res, cfg.LoggerOptions)
		handleErr(err, "failed to create the logger provider")
	} else {
		loggerProvider, err = newStdoutLoggerProvider(ctx, otelAgentAddr, res, cfg.LoggerOptions)
		handleErr(err, "failed to create the stdout logger provider")
	}

	// Set the logger provider globally
	global.SetLoggerProvider(loggerProvider)
	otelemetry.loggerProvider = loggerProvider
	otelemetry.logger = loggerProvider.Logger(serviceName, cfg.LoggerOptions.LoggerOption...)

	return &otelemetry, nil
}
