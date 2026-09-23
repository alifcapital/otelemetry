package otelemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
)

// Log interface provides methods for logging operations.
type Log interface {
	// Log returns the original logger.
	Log() log.Logger // original logger

	Debug(ctx context.Context, msg string, kv ...attribute.KeyValue)
	Info(ctx context.Context, msg string, kv ...attribute.KeyValue)
	Warning(ctx context.Context, msg string, kv ...attribute.KeyValue)
	Error(ctx context.Context, msg string, kv ...attribute.KeyValue)
	Fatal(ctx context.Context, msg string, kv ...attribute.KeyValue)
}

// otellog is an implementation of the Log interface using OpenTelemetry.
type otellog struct {
	log log.Logger
}

func newLoggerProvider(ctx context.Context, otelAgentAddr string, res *sdkresource.Resource, opts LoggerOptions) (*sdklog.LoggerProvider, error) {

	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
		otlploggrpc.WithEndpoint(otelAgentAddr),
	)
	if err != nil {
		return nil, err
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	return provider, nil
}

func newStdoutLoggerProvider(res *sdkresource.Resource) (*sdklog.LoggerProvider, error) {
	exporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}
	//stdoutlog.WithWriter(f),
	//stdoutlog.WithPrettyPrint(),

	provider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	)

	return provider, nil
}

const (
	Debug = "DEBUG"
	Info  = "INFO"
	Warn  = "WARN"
	Error = "ERROR"
	Fatal = "FATAL"
)

func (l *otellog) Log() log.Logger {
	return l.log
}

func (l *otellog) Debug(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	record := getRecord(msg, log.SeverityDebug, Debug, kv...)
	l.log.Emit(ctx, record)
}

func (l *otellog) Info(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	record := getRecord(msg, log.SeverityInfo, Info, kv...)
	l.log.Emit(ctx, record)
}

func (l *otellog) Warning(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	record := getRecord(msg, log.SeverityWarn, Warn, kv...)
	l.log.Emit(ctx, record)
}

func (l *otellog) Error(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	record := getRecord(msg, log.SeverityError, Error, kv...)
	l.log.Emit(ctx, record)
}

func (l *otellog) Fatal(ctx context.Context, msg string, kv ...attribute.KeyValue) {
	record := getRecord(msg, log.SeverityFatal, Fatal, kv...)
	l.log.Emit(ctx, record)
}

func getRecord(msg string, severity log.Severity, sevName string, kv ...attribute.KeyValue) log.Record {
	var record log.Record
	record.SetBody(attribute.StringValue(msg))
	record.SetSeverity(severity)
	record.SetSeverityText(sevName)
	record.SetTimestamp(time.Now())
	record.AddAttributes(kv...)
	return record
}
