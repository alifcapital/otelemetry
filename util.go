package otelemetry

import (
	"fmt"
	"math"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
)

func handleErr(err error, s string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", s, err))
	}
}

// Attribute builds an [attribute.KeyValue] from an arbitrary Go value for use
// in spans, metrics and log records. Unsupported types fall back to their
// fmt "%+v" string form.
func Attribute(k string, v any) attribute.KeyValue {
	return parseAttribute(k, v)
}

func parseAttribute(key string, value any) attribute.KeyValue {
	var attr attribute.KeyValue
	switch v := value.(type) {
	case string:
		attr = attribute.String(key, v)
	case []string:
		attr = attribute.StringSlice(key, v)
	case fmt.Stringer:
		attr = attribute.Stringer(key, v)
	case int:
		attr = attribute.Int(key, v)
	case []int:
		attr = attribute.IntSlice(key, v)
	case int8:
		attr = attribute.Int64(key, int64(v))
	case int16:
		attr = attribute.Int64(key, int64(v))
	case int32:
		attr = attribute.Int64(key, int64(v))
	case int64:
		attr = attribute.Int64(key, v)
	case []int64:
		attr = attribute.Int64Slice(key, v)
	case uint:
		attr = uintAttribute(key, uint64(v))
	case uint8:
		attr = attribute.Int64(key, int64(v))
	case uint16:
		attr = attribute.Int64(key, int64(v))
	case uint32:
		attr = attribute.Int64(key, int64(v))
	case uint64:
		attr = uintAttribute(key, v)
	case bool:
		attr = attribute.Bool(key, v)
	case []bool:
		attr = attribute.BoolSlice(key, v)
	case float32:
		attr = attribute.Float64(key, float64(v))
	case float64:
		attr = attribute.Float64(key, v)
	case []float64:
		attr = attribute.Float64Slice(key, v)
	default:
		attr = attribute.String(key, fmt.Sprintf("%+v", v))
	}
	return attr
}

// uintAttribute records an unsigned integer as an Int64 attribute when it fits,
// falling back to a decimal string when it would overflow int64 (OTel has no
// native unsigned attribute type).
func uintAttribute(key string, v uint64) attribute.KeyValue {
	if v <= math.MaxInt64 {
		return attribute.Int64(key, int64(v))
	}
	return attribute.String(key, strconv.FormatUint(v, 10))
}

// LogAttribute builds a log attribute from an arbitrary Go value.
//
// Deprecated: Use [Attribute] instead; log records take [attribute.KeyValue]
// since otel/log v0.21.0. LogAttribute will be removed in v0.4.0. To rewrite
// calls automatically, run
// `go run golang.org/x/tools/go/analysis/passes/inline/cmd/inline@latest -fix ./...`
// (see MIGRATION.md).
//
//go:fix inline
func LogAttribute(k string, v any) attribute.KeyValue {
	return Attribute(k, v)
}
