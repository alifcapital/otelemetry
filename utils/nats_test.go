package utils

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// mockJetstreamMsg implements jetstream.Msg for testing.
type mockJetstreamMsg struct {
	headers nats.Header
}

func (m *mockJetstreamMsg) Metadata() (*jetstream.MsgMetadata, error) { return nil, nil }
func (m *mockJetstreamMsg) Data() []byte                              { return nil }
func (m *mockJetstreamMsg) Headers() nats.Header                      { return m.headers }
func (m *mockJetstreamMsg) Subject() string                           { return "" }
func (m *mockJetstreamMsg) Reply() string                             { return "" }
func (m *mockJetstreamMsg) Ack() error                                { return nil }
func (m *mockJetstreamMsg) DoubleAck(context.Context) error           { return nil }
func (m *mockJetstreamMsg) Nak() error                                { return nil }
func (m *mockJetstreamMsg) NakWithDelay(time.Duration) error          { return nil }
func (m *mockJetstreamMsg) InProgress() error                         { return nil }
func (m *mockJetstreamMsg) Term() error                               { return nil }
func (m *mockJetstreamMsg) TermWithReason(string) error               { return nil }

func TestExtractsNatsTraceContextFromValidHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := nats.Msg{
		Header: nats.Header{
			"traceparent": []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			"tracestate":  []string{"key1=value1,key2=value2"},
		},
	}

	resultCtx := GetNatsTraceContext(ctx, msg)

	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", carrier["traceparent"])
	assert.Equal(t, "key1=value1,key2=value2", carrier["tracestate"])
}

func TestHandlesNatsHeadersWithEmptyValuesGracefully(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := nats.Msg{
		Header: nats.Header{
			"traceparent": []string{""},
		},
	}

	resultCtx := GetNatsTraceContext(ctx, msg)
	propagator := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	propagator.Inject(resultCtx, carrier)

	assert.Empty(t, carrier["traceparent"])
}

func TestInjectsTraceContextIntoNatsHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	propagator := propagation.TraceContext{}
	inCarrier := propagation.MapCarrier{
		"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"tracestate":  "key1=value1,key2=value2",
	}
	ctxWithTrace := propagator.Extract(ctx, inCarrier)

	headers := SetNatsHeaderTraceContext(ctxWithTrace)
	assert.Equal(t, inCarrier["traceparent"], headers.Get("traceparent"))
	assert.Equal(t, inCarrier["tracestate"], headers.Get("tracestate"))
}

func TestGetJetstreamTraceContext(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := &mockJetstreamMsg{
		headers: nats.Header{
			"traceparent": []string{"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"},
			"tracestate":  []string{"key1=value1"},
		},
	}

	resultCtx := GetJetstreamTraceContext(ctx, msg)

	prop := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	prop.Inject(resultCtx, carrier)

	assert.Equal(t, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", carrier["traceparent"])
	assert.Equal(t, "key1=value1", carrier["tracestate"])
}

func TestGetJetstreamTraceContextWithEmptyHeaders(t *testing.T) {
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx := context.Background()
	msg := &mockJetstreamMsg{headers: nats.Header{}}

	resultCtx := GetJetstreamTraceContext(ctx, msg)

	prop := propagation.TraceContext{}
	carrier := propagation.MapCarrier{}
	prop.Inject(resultCtx, carrier)

	assert.Empty(t, carrier["traceparent"])
}
