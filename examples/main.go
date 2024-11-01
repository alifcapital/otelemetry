package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/alifcapital/otelemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"
)

var telemetry otelemetry.Telemetry

func main() {

	ctx := context.Background()

	cfg := otelemetry.Config{
		Service: otelemetry.Service{
			Name:      "my-service",
			Namespace: "my-namespace",
			Version:   "1.0.0",
		},
		Collector: otelemetry.Collector{
			Host: "192.168.14.237",
			Port: "4317",
		},
		ResourceOptions: getResources(),
		WithMetrics:     false,
		WithLogs:        false,
		WithTraces:      false,
	}

	var err error
	telemetry, err = otelemetry.New(cfg)
	if err != nil {
		panic(err)
	}

	defer telemetry.Shutdown(ctx)

	mux := http.NewServeMux()
	mux.Handle("/hello", otelhttp.NewHandler(http.HandlerFunc(handler), "/hello"))
	mux.Handle("/hello-to-demo", otelhttp.NewHandler(http.HandlerFunc(handler2), "hello-to-demo"))
	server := &http.Server{
		Addr:              ":7080",
		Handler:           mux,
		ReadHeaderTimeout: 20 * time.Second,
	}
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func handler(w http.ResponseWriter, req *http.Request) {

	ctx := req.Context()
	ctx, span := telemetry.Trace().StartSpan(ctx, "handler", trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	sleep := rng.Int63n(1000)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	requestCount, err := telemetry.Metric().Float64Counter("example_request_count")
	if err != nil {
		panic(err)
	}

	requestCount.Add(ctx, 1)

	telemetry.Log().Info(ctx, "Request processed", otelemetry.LogAttribute("sleep", sleep))

	span.AddEvent("sleep event", otelemetry.Attribute("sleep", sleep))

	// do some tracing with new span on other func
	doTraceWithNewSpan(ctx)

	// do some tracing with current span on other func
	doTraceWithCurrentSpan(ctx)

	if _, err := w.Write([]byte(fmt.Sprintf("Sleep: %dms", sleep))); err != nil {
		http.Error(w, "write operation failed.", http.StatusInternalServerError)
		return
	}
}

func doTraceWithNewSpan(ctx context.Context) {
	ctx, span := telemetry.Trace().StartSpan(ctx, "new span")
	defer span.End()

	sleep := rng.Int63n(100)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	span.AddEvent("event: trace with new span", otelemetry.Attribute("sleep", sleep))
}

func doTraceWithCurrentSpan(ctx context.Context) {
	span := telemetry.Trace().SpanFromContext(ctx)
	if span == nil {
		telemetry.Log().Error(ctx, "failed to get span from context")
		return
	}

	sleep := rng.Int63n(100)
	time.Sleep(time.Duration(sleep) * time.Millisecond)

	span.AddEvent("event: trace with current span", otelemetry.Attribute("sleep", sleep))
}

func getResources() []resource.Option {
	return []resource.Option{
		resource.WithHost(),
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
	}
}

func handler2(w http.ResponseWriter, req *http.Request) {

	ctx, span := telemetry.Trace().StartSpan(context.Background(), "ExecuteRequest")
	makeRequest(ctx)
	span.End()

	if _, err := w.Write([]byte(fmt.Sprintf("Request send"))); err != nil {
		http.Error(w, "write operation failed.", http.StatusInternalServerError)
		return
	}
}

func makeRequest(ctx context.Context) {

	demoServerAddr := "http://localhost:9006/api/v1/hello-example"

	// Trace an HTTP client by wrapping the transport
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	// Make sure we pass the context to the request to avoid broken traces.
	req, err := http.NewRequestWithContext(ctx, "GET", demoServerAddr, nil)
	if err != nil {
		panic(err)
	}

	// All requests made with this client will create spans.
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	res.Body.Close()
}

//	carrier := propagation.MapCarrier{}
//	propagator := otel.GetTextMapPropagator()
//	propagator.Inject(spanContext, carrier)
//
//	for key, value := range carrier {
//		msg.Headers = append(msg.Headers, sarama.RecordHeader{Key: []byte(key), Value: []byte(value)})
//	}
