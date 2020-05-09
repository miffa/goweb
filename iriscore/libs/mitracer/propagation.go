package mitracer

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/opentracing/opentracing-go"
)

const miTextMapIdsPrefix = "mipfx-ids-"
const miTextMapBaggagePrefix = "mipfx-baggage-"

var emptyContext = MiSpanContext{}

// Injector is responsible for injecting SpanContext instances in a manner suitable
// for propagation via a format-specific "carrier" object. Typically the
// injection will take place across an RPC boundary, but message queues and
// other IPC mechanisms are also reasonable places to use an Injector.
type Injector interface {
	// Inject takes `SpanContext` and injects it into `carrier`. The actual type
	// of `carrier` depends on the `format` passed to `Tracer.Inject()`.
	//
	// Implementations may return opentracing.ErrInvalidCarrier or any other
	// implementation-specific error if injection fails.
	Inject(ctx MiSpanContext, carrier interface{}) error
}

// Extractor is responsible for extracting SpanContext instances from a
// format-specific "carrier" object. Typically the extraction will take place
// on the server side of an RPC boundary, but message queues and other IPC
// mechanisms are also reasonable places to use an Extractor.
type Extractor interface {
	// Extract decodes a SpanContext instance from the given `carrier`,
	// or (nil, opentracing.ErrSpanContextNotFound) if no context could
	// be found in the `carrier`.
	Extract(carrier interface{}) (MiSpanContext, error)
}

// TextMapPropagator implements Injector/Extractor for TextMap and HTTPHeaders formats.
type TextMapPropagator struct {
	HTTPHeaders bool
}

// Inject implements the Injector interface
func (t *TextMapPropagator) Inject(spanContext MiSpanContext, carrier interface{}) error {
	writer, ok := carrier.(opentracing.TextMapWriter)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}
	// Ids:
	writer.Set(miTextMapIdsPrefix+"traceid", strconv.FormatUint(spanContext.TraceID, 10))
	writer.Set(miTextMapIdsPrefix+"spanid", strconv.FormatUint(spanContext.SpanID, 10))
	writer.Set(miTextMapIdsPrefix+"sampled", fmt.Sprint(spanContext.Sampled))
	// Baggage:
	for baggageKey, baggageVal := range spanContext.Baggage {
		safeVal := baggageVal
		if t.HTTPHeaders {
			safeVal = url.QueryEscape(baggageVal)
		}
		writer.Set(miTextMapBaggagePrefix+baggageKey, safeVal)
	}
	return nil
}

// Extract implements the Extractor interface
func (t *TextMapPropagator) Extract(carrier interface{}) (MiSpanContext, error) {
	reader, ok := carrier.(opentracing.TextMapReader)
	if !ok {
		return emptyContext, opentracing.ErrInvalidCarrier
	}
	rval := MiSpanContext{0, 0, true, nil}
	err := reader.ForeachKey(func(key, val string) error {
		lowerKey := strings.ToLower(key)
		switch {
		case lowerKey == miTextMapIdsPrefix+"traceid":
			// Ids:
			i, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}
			rval.TraceID = i
		case lowerKey == miTextMapIdsPrefix+"spanid":
			// Ids:
			i, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return err
			}
			rval.SpanID = i
		case lowerKey == miTextMapIdsPrefix+"sampled":
			b, err := strconv.ParseBool(val)
			if err != nil {
				return err
			}
			rval.Sampled = b
		case strings.HasPrefix(lowerKey, miTextMapBaggagePrefix):
			// Baggage:
			if rval.Baggage == nil {
				rval.Baggage = make(map[string]string)
			}
			safeVal := val
			if t.HTTPHeaders {
				// unescape errors are ignored, nothing can be done
				if rawVal, err := url.QueryUnescape(val); err == nil {
					safeVal = rawVal
				}
			}
			rval.Baggage[lowerKey[len(miTextMapBaggagePrefix):]] = safeVal
		}
		return nil
	})
	if rval.TraceID == 0 || rval.SpanID == 0 {
		return emptyContext, opentracing.ErrSpanContextNotFound
	}
	if err != nil {
		return emptyContext, err
	}
	return rval, nil
}
