# Migration guide

## v0.2.x → v0.3.0

`LogAttribute` is **deprecated** and will be **removed in v0.4.0**. Use `Attribute` instead. Since v0.2.0, both return `attribute.KeyValue`, so `Attribute` works for spans, metrics and logs.

### Automatic rewrite

`LogAttribute` carries a `//go:fix inline` directive, so its calls can be rewritten mechanically:

```bash
go run golang.org/x/tools/go/analysis/passes/inline/cmd/inline@latest -fix ./...
```

This turns `otelemetry.LogAttribute(k, v)` into `otelemetry.Attribute(k, v)`.

The built-in `go fix ./...` in Go 1.26.x also rewrites the calls, but it leaves an unused `go.opentelemetry.io/otel/attribute` import behind, so the code won't compile. If you use it, run `goimports -w .` afterwards.

Until you migrate, staticcheck (SA1019), gopls and IDEs flag every remaining call.

### Behavior change

In v0.3.0, `LogAttribute` is an exact alias of `Attribute`. For most types the result is unchanged. The types below used to be stringified and are now recorded with their native type:

| Value type | v0.2.x `LogAttribute` | v0.3.0 (`Attribute`) |
|------------|-----------------------|----------------------|
| `int8/16/32`, `uint*` | `STRING "1"` | `INT64 1` (a `uint64` above `MaxInt64` stays a decimal string) |
| `float32` | `STRING "1.5"` | `FLOAT64 1.5` |
| `[]string`, `[]int`, `[]int64`, `[]bool`, `[]float64` | `STRING "[a b]"` | typed slice, e.g. `STRINGSLICE ["a","b"]` |

If your log queries, alerts or dashboards filter on such fields as strings, update them.

## v0.1.x → v0.2.0

v0.2.0 upgrades OpenTelemetry from v1.44.0 to v1.46.0 and the logs modules from v0.20.0 to v0.22.0. It also picks up security fixes in gRPC, `golang.org/x/crypto`, `x/net`, `x/text` and `klauspost/compress`.

Starting with `go.opentelemetry.io/otel/log` v0.21.0, log attributes use the shared attribute types. `log.KeyValue`, `log.Value`, `log.Kind` and their constructors (`log.String`, `log.Int64`, …) no longer exist ([open-telemetry/opentelemetry-go#8490](https://github.com/open-telemetry/opentelemetry-go/pull/8490)). This library follows that change.

### API changes

| v0.1.x | v0.2.0 |
|--------|--------|
| `Log.Debug/Info/Warning/Error/Fatal(ctx, msg, kv ...log.KeyValue)` | `Log.Debug/Info/Warning/Error/Fatal(ctx, msg, kv ...attribute.KeyValue)` |
| `LogAttribute(k, v) log.KeyValue` | `LogAttribute(k, v) attribute.KeyValue` |

`attribute` is `go.opentelemetry.io/otel/attribute`. `LogAttribute` still maps the same Go types to the same value kinds.

### Do I need to change anything?

**No**, if you only pass `otelemetry.LogAttribute(...)` inline:

```go
tel.Log().Info(ctx, "user signed in", otelemetry.LogAttribute("user_id", userID))
```

**Yes**, in these cases:

1. **You pass `go.opentelemetry.io/otel/log` constructors.** Switch to `attribute`:

   ```go
   // before
   tel.Log().Info(ctx, "msg", log.String("k", "v"), log.Int64("n", 1))
   // after
   tel.Log().Info(ctx, "msg", attribute.String("k", "v"), attribute.Int64("n", 1))
   ```

2. **You store log attributes in typed variables, slices or fields.** Replace `log.KeyValue` with `attribute.KeyValue`:

   ```go
   // before
   attrs := []log.KeyValue{otelemetry.LogAttribute("k", "v")}
   // after
   attrs := []attribute.KeyValue{otelemetry.LogAttribute("k", "v")}
   ```

3. **You implement or mock the `otelemetry.Log` interface.** Update the method signatures to `kv ...attribute.KeyValue`.

4. **Your project uses the `otel/log` API directly or through a bridge** such as `otelslog`, `otelzap` or `otellogrus`. Go picks a single version of a module for the whole build, so your project also moves to `otel/log` v0.22.0. Upgrade the bridges to a release built against v0.22.0. For `go.opentelemetry.io/contrib/bridges/*` that is v0.20.1 or later.

### Behavior changes

- **stdout log exporter** (used when `WithLogs: false`): log bodies and attributes are now encoded as `attribute.Value` JSON, for example `{"Type":"STRING","Value":"..."}`. Update anything that parses this output. The OTLP wire format sent to the collector is unchanged.
- See the upstream [OpenTelemetry-Go changelog](https://github.com/open-telemetry/opentelemetry-go/blob/main/CHANGELOG.md) for v1.45.0 and v1.46.0 for SDK-level changes.

### Toolchain

Several standard library vulnerabilities (`crypto/tls`, `encoding/asn1`, `net/http`, …) are fixed only in **Go 1.26.6**. Build with Go 1.26.6 or newer.
