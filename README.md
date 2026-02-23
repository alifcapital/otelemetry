# OTelemetry

OTelemetry is a wrapper around the [OpenTelemetry](https://opentelemetry.io/) Go SDK that provides a simple interface to instrument your code with traces, metrics, and logs — exported over gRPC to an OpenTelemetry Collector.

When the `With*` flags are disabled the library falls back to stdout exporters, which is convenient for local development and debugging.

### Install

```shell
go get -u github.com/alifcapital/otelemetry
```

### Initialization

```go
import "github.com/alifcapital/otelemetry"

func main() {
    cfg := otelemetry.Config{
        Service: otelemetry.Service{
            Name:      "example-service",
            Namespace: "payments",
            Version:   "1.0.0",
        },
        Collector: otelemetry.Collector{
            Host: "localhost",
            Port: "4317",
        },
        WithTraces:  true,
        WithMetrics: true,
        WithLogs:    true,
    }

    tel, err := otelemetry.New(cfg)
    if err != nil {
        log.Fatalf("failed to initialize telemetry: %v", err)
    }
    defer tel.Shutdown(context.Background())
}
```

Advanced options are passed through the provider-specific option structs. For example, to enable gzip compression for traces:

```go
import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"

cfg := otelemetry.Config{
    // ...
    TracerOptions: otelemetry.TracerOptions{
        ClientOption: []otlptracegrpc.Option{
            otlptracegrpc.WithCompressor("gzip"),
        },
    },
    MetricOptions: otelemetry.MetricOptions{
        PeriodicInterval: 10 * time.Second, // default: 5s
    },
}
```

### Tracing

```go
// Start a new span
ctx, span := tel.Trace().StartSpan(context.Background(), "operation-name")
defer span.End()

// Add an event
span.AddEvent("something happened", otelemetry.Attribute("key", "value"))

// Add a structured error event (sets span status to Error)
span.AddErrorEvent("db query failed", err, otelemetry.Attribute("query", sql))

// Record an error
span.RecordError(err)

// Set attributes on the span
span.SetAttribute(otelemetry.Attribute("user.id", userID))

// Get trace/span IDs (useful for logging correlation)
traceID := span.TraceID()
spanID  := span.SpanID()
```

**Continuing a span from context:**

```go
span := tel.Trace().SpanFromContext(ctx)
span.AddEvent("resumed span")
```

### Metrics

All standard OpenTelemetry instrument types are supported:

```go
// Counters
counter, err := tel.Metric().Float64Counter("requests_total")
counter.Add(ctx, 1, metric.WithAttributes(otelemetry.Attribute("status", "ok")))

intCounter, err := tel.Metric().Int64Counter("errors_total")

// Gauges
gauge, err := tel.Metric().Float64Gauge("queue_depth")
gauge.Record(ctx, float64(len(queue)))

// Histograms
hist, err := tel.Metric().Float64Histogram("request_duration_seconds")
hist.Record(ctx, duration.Seconds())

// Up-down counters
udCounter, err := tel.Metric().Int64UpDownCounter("active_connections")

// Observable (async) instruments
obsGauge, err := tel.Metric().Float64ObservableGauge("memory_usage_bytes")
tel.Metric().RegisterCallback(func(ctx context.Context, o metric.Observer) error {
    o.ObserveFloat64(obsGauge, getMemoryUsage())
    return nil
}, obsGauge)
```

### Logging

```go
tel.Log().Debug(ctx, "debug message", otelemetry.LogAttribute("key", "value"))
tel.Log().Info(ctx, "user signed in", otelemetry.LogAttribute("user_id", userID))
tel.Log().Warning(ctx, "slow query", otelemetry.LogAttribute("duration_ms", 450))
tel.Log().Error(ctx, "payment failed", otelemetry.LogAttribute("error", err.Error()))
tel.Log().Fatal(ctx, "unrecoverable failure")
```

### Context Propagation

**HTTP:**

```go
// Inject the current span context into outgoing HTTP request headers
otelemetry.InjectHTTPHeaders(ctx, req.Header)

// Extract span context from incoming HTTP request headers
ctx = otelemetry.ExtractHTTPHeaders(ctx, r.Header)
```

**Generic map carrier (gRPC metadata, custom protocols, etc.):**

```go
// Inject
headers := make(map[string]string)
otelemetry.Inject(ctx, headers)

// Extract
ctx = otelemetry.Extract(ctx, headers)
```

### Message Queue Utilities

Import the `utils` subpackage:

```go
import otelutils "github.com/alifcapital/otelemetry/utils"
```

**NATS / JetStream:**

```go
// Consumer: extract trace context from incoming message
func (h *handler) SignedIn(msg jetstream.Msg) {
    ctx := otelutils.GetJetstreamTraceContext(context.Background(), msg)
    ctx, span := tel.Trace().StartSpan(ctx, "NatsHandler: user.SignedIn")
    defer span.End()
    // ...
}

// For plain nats.Msg:
ctx = otelutils.GetNatsTraceContext(context.Background(), *msg)

// Producer: inject trace context into outgoing message headers
header := otelutils.SetNatsHeaderTraceContext(ctx)
msg := &nats.Msg{Subject: "user.signed_in", Header: header}
nc.PublishMsg(msg)
```

**RabbitMQ:**

```go
// Consumer: extract trace context from incoming delivery
func (h *handler) handle(d amqp.Delivery) {
    ctx := otelutils.GetRabbitMQTraceContext(context.Background(), d)
    ctx, span := tel.Trace().StartSpan(ctx, "RabbitMQ: user.SignedIn")
    defer span.End()
    // ...
}

// Producer: inject trace context into publishing headers
headers := otelutils.SetRabbitMQHeaderTraceContext(ctx)
ch.PublishWithContext(ctx, exchange, key, false, false, amqp.Publishing{
    Headers: headers,
    Body:    body,
})
```

### Baggage

```go
// Add baggage items to context
ctx = otelemetry.AddBaggageItem(ctx, "tenant_id", "acme")
ctx = otelemetry.AddBaggageItems(ctx, map[string]string{"env": "prod", "region": "eu-west"})

// Read baggage
tenantID := otelemetry.GetBaggageItem(ctx, "tenant_id")
allBaggage := otelemetry.GetBaggage(ctx)

// Remove a baggage item
ctx = otelemetry.RemoveBaggageItem(ctx, "tenant_id")
```

### Contributing

Pull requests are welcome.
