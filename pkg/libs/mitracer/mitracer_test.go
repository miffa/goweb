package mitracer

import (
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

func TestMiTracer_StartSpan(t *testing.T) {
	tracer := New()
	span1 := tracer.StartSpan(
		"a",
		opentracing.Tags(map[string]interface{}{"x": "y"}))

	span2 := span1.Tracer().StartSpan(
		"", opentracing.ChildOf(span1.Context()))
	span2.Finish()
	span1.Finish()
	spans := tracer.FinishedSpans()
	assert.Equal(t, 2, len(spans))

	parent := spans[1]
	child := spans[0]
	assert.Equal(t, map[string]interface{}{"x": "y"}, parent.Tags())
	assert.Equal(t, child.ParentID, parent.Context().(MiSpanContext).SpanID)
}

func TestMiSpan_SetOperationName(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("")
	span.SetOperationName("x")
	assert.Equal(t, "x", span.(*MiSpan).OperationName)
}

func TestMiSpanContext_Baggage(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("x")
	span.SetBaggageItem("x", "y")
	assert.Equal(t, "y", span.BaggageItem("x"))
	assert.Equal(t, map[string]string{"x": "y"}, span.Context().(MiSpanContext).Baggage)

	baggage := make(map[string]string)
	span.Context().ForeachBaggageItem(func(k, v string) bool {
		baggage[k] = v
		return true
	})
	assert.Equal(t, map[string]string{"x": "y"}, baggage)

	span.SetBaggageItem("a", "b")
	baggage = make(map[string]string)
	span.Context().ForeachBaggageItem(func(k, v string) bool {
		baggage[k] = v
		return false // exit early
	})
	assert.Equal(t, 2, len(span.Context().(MiSpanContext).Baggage))
	assert.Equal(t, 1, len(baggage))
}

func TestMiSpan_Tag(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("x")
	span.SetTag("x", "y")
	assert.Equal(t, "y", span.(*MiSpan).Tag("x"))
}

func TestMiSpan_Tags(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("x")
	span.SetTag("x", "y")
	assert.Equal(t, map[string]interface{}{"x": "y"}, span.(*MiSpan).Tags())
}

func TestMiTracer_FinishedSpans_and_Reset(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("x")
	span.SetTag("x", "y")
	span.Finish()
	spans := tracer.FinishedSpans()
	assert.Equal(t, 1, len(spans))
	assert.Equal(t, map[string]interface{}{"x": "y"}, spans[0].Tags())

	tracer.Reset()
	spans = tracer.FinishedSpans()
	assert.Equal(t, 0, len(spans))
}

func zeroOutTimestamps(recs []MiLogRecord) {
	for i := range recs {
		recs[i].Timestamp = time.Time{}
	}
}

func TestMiSpan_LogKV(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("s")
	span.LogKV("key0", "string0")
	span.LogKV("key1", "string1", "key2", uint32(42))
	span.Finish()
	spans := tracer.FinishedSpans()
	assert.Equal(t, 1, len(spans))
	actual := spans[0].Logs()
	zeroOutTimestamps(actual)
	assert.Equal(t, []MiLogRecord{
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "key0", ValueKind: reflect.String, ValueString: "string0"},
			},
		},
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "key1", ValueKind: reflect.String, ValueString: "string1"},
				MiKeyValue{Key: "key2", ValueKind: reflect.Uint32, ValueString: "42"},
			},
		},
	}, actual)
}

func TestMiSpan_LogFields(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("s")
	span.LogFields(log.String("key0", "string0"))
	span.LogFields(log.String("key1", "string1"), log.Uint32("key2", uint32(42)))
	span.LogFields(log.Lazy(func(fv log.Encoder) {
		fv.EmitInt("key_lazy", 12)
	}))
	span.FinishWithOptions(opentracing.FinishOptions{
		LogRecords: []opentracing.LogRecord{
			{Timestamp: time.Now(), Fields: []log.Field{log.String("key9", "finish")}},
		}})
	spans := tracer.FinishedSpans()
	assert.Equal(t, 1, len(spans))
	actual := spans[0].Logs()
	zeroOutTimestamps(actual)
	assert.Equal(t, []MiLogRecord{
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "key0", ValueKind: reflect.String, ValueString: "string0"},
			},
		},
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "key1", ValueKind: reflect.String, ValueString: "string1"},
				MiKeyValue{Key: "key2", ValueKind: reflect.Uint32, ValueString: "42"},
			},
		},
		MiLogRecord{
			Fields: []MiKeyValue{
				// Note that the LazyLogger gets to control the key as well as the value.
				MiKeyValue{Key: "key_lazy", ValueKind: reflect.Int, ValueString: "12"},
			},
		},
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "key9", ValueKind: reflect.String, ValueString: "finish"},
			},
		},
	}, actual)
}

func TestMiSpan_DeprecatedLogs(t *testing.T) {
	tracer := New()
	span := tracer.StartSpan("x")
	span.LogEvent("x")
	span.LogEventWithPayload("y", "z")
	span.LogEvent("a")
	span.FinishWithOptions(opentracing.FinishOptions{
		BulkLogData: []opentracing.LogData{{Event: "f"}}})
	spans := tracer.FinishedSpans()
	assert.Equal(t, 1, len(spans))
	actual := spans[0].Logs()
	zeroOutTimestamps(actual)
	assert.Equal(t, []MiLogRecord{
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "event", ValueKind: reflect.String, ValueString: "x"},
			},
		},
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "event", ValueKind: reflect.String, ValueString: "y"},
				MiKeyValue{Key: "payload", ValueKind: reflect.String, ValueString: "z"},
			},
		},
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "event", ValueKind: reflect.String, ValueString: "a"},
			},
		},
		MiLogRecord{
			Fields: []MiKeyValue{
				MiKeyValue{Key: "event", ValueKind: reflect.String, ValueString: "f"},
			},
		},
	}, actual)
}

func TestMiTracer_Propagation(t *testing.T) {
	textCarrier := func() interface{} {
		return opentracing.TextMapCarrier(make(map[string]string))
	}
	textLen := func(c interface{}) int {
		return len(c.(opentracing.TextMapCarrier))
	}

	httpCarrier := func() interface{} {
		httpHeaders := http.Header(make(map[string][]string))
		return opentracing.HTTPHeadersCarrier(httpHeaders)
	}
	httpLen := func(c interface{}) int {
		return len(c.(opentracing.HTTPHeadersCarrier))
	}

	tests := []struct {
		sampled bool
		format  opentracing.BuiltinFormat
		carrier func() interface{}
		len     func(interface{}) int
	}{
		{sampled: true, format: opentracing.TextMap, carrier: textCarrier, len: textLen},
		{sampled: false, format: opentracing.TextMap, carrier: textCarrier, len: textLen},
		{sampled: true, format: opentracing.HTTPHeaders, carrier: httpCarrier, len: httpLen},
		{sampled: false, format: opentracing.HTTPHeaders, carrier: httpCarrier, len: httpLen},
	}
	for _, test := range tests {
		tracer := New()
		span := tracer.StartSpan("x")
		span.SetBaggageItem("x", "y:z") // colon should be URL encoded as %3A
		if !test.sampled {
			ext.SamplingPriority.Set(span, 0)
		}
		mSpan := span.(*MiSpan)

		assert.Equal(t, opentracing.ErrUnsupportedFormat,
			tracer.Inject(span.Context(), opentracing.Binary, nil))
		assert.Equal(t, opentracing.ErrInvalidCarrier,
			tracer.Inject(span.Context(), opentracing.TextMap, span))

		carrier := test.carrier()

		err := tracer.Inject(span.Context(), test.format, carrier)
		require.NoError(t, err)
		assert.Equal(t, 4, test.len(carrier), "expect baggage + 2 ids + sampled")
		if test.format == opentracing.HTTPHeaders {
			c := carrier.(opentracing.HTTPHeadersCarrier)
			assert.Equal(t, "y%3Az", c["Mipfx-Baggage-X"][0])
		}

		_, err = tracer.Extract(opentracing.Binary, nil)
		assert.Equal(t, opentracing.ErrUnsupportedFormat, err)
		_, err = tracer.Extract(opentracing.TextMap, tracer)
		assert.Equal(t, opentracing.ErrInvalidCarrier, err)

		extractedContext, err := tracer.Extract(test.format, carrier)
		require.NoError(t, err)
		assert.Equal(t, mSpan.SpanContext.TraceID, extractedContext.(MiSpanContext).TraceID)
		assert.Equal(t, mSpan.SpanContext.SpanID, extractedContext.(MiSpanContext).SpanID)
		assert.Equal(t, test.sampled, extractedContext.(MiSpanContext).Sampled)
		assert.Equal(t, "y:z", extractedContext.(MiSpanContext).Baggage["x"])
	}
}

func TestMiSpan_Races(t *testing.T) {
	span := New().StartSpan("x")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		span.SetBaggageItem("test_key", "test_value")
	}()
	go func() {
		defer wg.Done()
		span.Context()
	}()
	wg.Wait()
}
