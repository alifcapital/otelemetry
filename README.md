# OTelemetry

OTelemetry is a wrapper around (over gRPC) the [OpenTelemetry](https://opentelemetry.io/) library
that provides a simple interface to instrument your code with telemetry.

### ToDo
- [ ] Jetstream utils
- [ ] RabbitMQ utils
- [ ] TLS support
- [ ] Add tests
- [ ] Modified examples
- [ ] Add documentation

### Install

```shell
go get -u github.com/alifcapital/otelemetry
```


### Usage

Here is a simple example of how to use OTelemetry in your Go application:

```go
import "github.com/alifcapital/otelemetry"

func main() {
	// Configuration for OTelemetry
	cfg := otelemetry.Config{
		Service: otelemetry.ServiceConfig{
			Name: "example-service",
		},
		Collector: otelemetry.CollectorConfig{
			Host: "localhost",
			Port: "4317",
		},
                TracerOptions: otelemetry.TracerOptions{
                    ClientOption: []otlptracegrpc.Option{
                        otlptracegrpc.WithCompressor("gzip"),
                    },
                },
		WithTraces:  true,
		WithMetrics: true,
		WithLogs:    true,
	}

	// Initialize OTelemetry
	tel, err := otelemetry.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize telemetry: %v", err)
	}
	defer tel.Shutdown(context.Background())
	
	// your code 
}	
```

Example usage of tracer and span:
```go
// Example usage of tracer and span
ctx, span := tel.Trace().StartSpan(context.Background(), "example-span")
defer span.End()

span.AddEvent("example event", otelemetry.Attribute("key", "value"))
```

Get span from context:
```go
span := tel.Trace().SpanFromContext(ctx)

span.AddEvent("example of continuing span get from context", otelemetry.Attribute("key", "value"))
```

Example usage meter:

```go
// Example usage meter
counter, err := tel.Metric().Float64Counter("example_counter")
if err != nil {
    panic(err)
}

counter.Add(ctx, 1)
```

Example usage of logger:

```go
// Example usage of logger
tel.Log().Info(ctx, "log message", otelemetry.LogAttribute("key", "value"))
```


Example of getting a context with tracing data from Nats message:

```go

import (
    otelemetryutils "github.com/alifcapital/otelemetry/utils"
)

func (h *handler) SignedIn(msg jetstream.Msg) {
    
    ctx := otelemetryutils.GetNatsTraceContext(context.Background(), *msg)
    
    ctx, span := tel.Trace().StartSpan(ctx, "NatsHandler: user.SignedIn")
    defer span.End()
    
    // code ...
}

```

### Contributing

Pull requests are welcome.
