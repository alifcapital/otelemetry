# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests (root + utils packages)
go test ./...

# Run tests for a specific package
go test ./utils/...

# Run a single test
go test -run TestName ./...

# Verbose test output
go test -v ./...

# Build all packages
go build ./...

# Tidy dependencies
go mod tidy
```

There is no Makefile. The module is `github.com/alifcapital/otelemetry` (Go 1.26).

## Architecture

This library is a **thin wrapper over OpenTelemetry** that simplifies instrumentation setup. When `WithTraces/WithMetrics/WithLogs` is `true`, it exports via OTLP gRPC to a collector; when `false`, it falls back to stdout exporters (for local dev/debug).

### Package layout

- **Root package (`otelemetry`)** — Public API. All user-facing interfaces live here.
- **`utils/`** — Trace context propagation helpers for message queues (NATS/JetStream and RabbitMQ). Each provides `Get*TraceContext` (extract from incoming message headers into `context.Context`) and `Set*HeaderTraceContext` (inject current span into outgoing message headers).

### Key interfaces and their implementations

| Interface | Impl struct | File |
|-----------|-------------|------|
| `Telemetry` | `telemetry` | `otel.go` |
| `Trace` | `oteltrace` | `trace.go` |
| `Span` | `otelspan` | `span.go` |
| `Metric` | `otelmetric` | `metrics.go` |
| `Log` | `otellog` | `logs.go` |

All implementation structs are unexported; only interfaces are public.

### Configuration (`structs.go`)

`Config` is the single configuration struct passed to `New()`:
- `Service{Name, Namespace, Version}` — resource identification
- `Collector{Host, Port}` — OTLP gRPC endpoint (e.g., `"localhost"`, `"4317"`)
- `WithTraces / WithMetrics / WithLogs bool` — toggle OTLP export vs stdout fallback
- `TracerOptions`, `MetricOptions`, `LoggerOptions` — pass-through option slices for the underlying SDK (exporter, provider, tracer/meter/logger options)
- `MetricOptions.PeriodicInterval` — controls the periodic reader interval (default: 5s)
- `ResourceOptions []sdkresource.Option` — additional OTel resource attributes

### Initialization flow (`New()` in `otel.go`)

1. Build an OTel `Resource` via `newResource()` (uses `Service` fields + any `ResourceOptions`)
2. Create tracer/meter/logger providers (OTLP or stdout depending on `With*` flags)
3. Register each provider globally via `otel.Set*Provider` / `global.SetLoggerProvider`
4. Set the global propagator to `TraceContext + Baggage` (W3C)

### Helper functions

- `Attribute(k, v any)` / `LogAttribute(k, v any)` — type-safe `attribute.KeyValue` / `log.KeyValue` builders (`util.go`)
- `Inject` / `Extract` / `InjectHTTPHeaders` / `ExtractHTTPHeaders` — W3C trace context propagation over map/HTTP carriers (`propagation.go`)
- Baggage helpers: `GetBaggage`, `AddBaggageItem`, `AddBaggageItems`, `GetBaggageItem`, `RemoveBaggageItem` (`baggage.go`)

### Go workspace

`go.work` ties the root module and `examples/` together. The `examples/` directory has its own `go.mod` and is a standalone runnable HTTP server demonstrating typical usage.
