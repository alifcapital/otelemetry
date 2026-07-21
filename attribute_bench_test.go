package otelemetry

import (
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// sink prevents the compiler from optimizing benchmarked calls away.
var (
	sinkKV  attribute.KeyValue
	sinkKVs []attribute.KeyValue
)

// --- Individual types through Attribute (any -> type switch) ---

func BenchmarkAttribute_String(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("http.method", "GET")
	}
}

func BenchmarkAttribute_Int(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("http.status_code", 200)
	}
}

func BenchmarkAttribute_Int64(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("bytes", int64(4096))
	}
}

func BenchmarkAttribute_Float64(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("latency_ms", 12.5)
	}
}

func BenchmarkAttribute_Bool(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("cache.hit", true)
	}
}

// Newly covered numeric types (previously hit the fmt.Sprintf default branch).

func BenchmarkAttribute_Int32(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("shard", int32(12))
	}
}

func BenchmarkAttribute_Uint(b *testing.B) {
	var v uint = 42
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("count", v)
	}
}

func BenchmarkAttribute_Uint64(b *testing.B) {
	var v uint64 = 4096
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("bytes", v)
	}
}

func BenchmarkAttribute_Float32(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("ratio", float32(0.5))
	}
}

// --- Direct OTel constructors (no any boxing, no type switch) ---

func BenchmarkDirect_String(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = attribute.String("http.method", "GET")
	}
}

func BenchmarkDirect_Int(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = attribute.Int("http.status_code", 200)
	}
}

func BenchmarkDirect_Int64(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = attribute.Int64("bytes", 4096)
	}
}

func BenchmarkDirect_Float64(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = attribute.Float64("latency_ms", 12.5)
	}
}

func BenchmarkDirect_Bool(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = attribute.Bool("cache.hit", true)
	}
}

// --- The expensive default branch: fmt.Sprintf("%#v", v) + reflection ---

type customType struct {
	ID   int
	Name string
}

func BenchmarkAttribute_DefaultBranch_Struct(b *testing.B) {
	v := customType{ID: 7, Name: "abc"}
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("payload", v)
	}
}

// Stringer branch: v.String() typically allocates.
func BenchmarkAttribute_Stringer_Time(b *testing.B) {
	v := time.Unix(1700000000, 0)
	b.ReportAllocs()
	for b.Loop() {
		sinkKV = Attribute("ts", v)
	}
}

// --- Realistic hot path: building a set of span attributes per request ---

func BenchmarkAttribute_RequestSet(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKVs = []attribute.KeyValue{
			Attribute("http.method", "GET"),
			Attribute("http.route", "/api/v1/users"),
			Attribute("http.status_code", 200),
			Attribute("http.request_bytes", int64(1024)),
			Attribute("cache.hit", true),
		}
	}
}

func BenchmarkDirect_RequestSet(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sinkKVs = []attribute.KeyValue{
			attribute.String("http.method", "GET"),
			attribute.String("http.route", "/api/v1/users"),
			attribute.Int("http.status_code", 200),
			attribute.Int64("http.request_bytes", 1024),
			attribute.Bool("cache.hit", true),
		}
	}
}
