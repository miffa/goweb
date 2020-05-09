package mitracer

import (
	"sync"

	"github.com/opentracing/opentracing-go"
)

// New returns a MiTracer opentracing.Tracer implementation that's intended
// to facilitate tests of OpenTracing instrumentation.
func New() *MiTracer {
	t := &MiTracer{
		finishedSpans: []*MiSpan{},
		injectors:     make(map[interface{}]Injector),
		extractors:    make(map[interface{}]Extractor),
	}

	// register default injectors/extractors
	textPropagator := new(TextMapPropagator)
	t.RegisterInjector(opentracing.TextMap, textPropagator)
	t.RegisterExtractor(opentracing.TextMap, textPropagator)

	httpPropagator := &TextMapPropagator{HTTPHeaders: true}
	t.RegisterInjector(opentracing.HTTPHeaders, httpPropagator)
	t.RegisterExtractor(opentracing.HTTPHeaders, httpPropagator)

	return t
}

// MiTracer is only intended for testing OpenTracing instrumentation.
//
// It is entirely unsuitable for production use, but appropriate for tests
// that want to verify tracing behavior in other frameworks/applications.
type MiTracer struct {
	sync.RWMutex
	finishedSpans []*MiSpan
	injectors     map[interface{}]Injector
	extractors    map[interface{}]Extractor
}

// FinishedSpans returns all spans that have been Finish()'ed since the
// MiTracer was constructed or since the last call to its Reset() method.
func (t *MiTracer) FinishedSpans() []*MiSpan {
	t.RLock()
	defer t.RUnlock()
	spans := make([]*MiSpan, len(t.finishedSpans))
	copy(spans, t.finishedSpans)
	return spans
}

// Reset clears the internally accumulated finished spans. Note that any
// extant MiSpans will still append to finishedSpans when they Finish(),
// even after a call to Reset().
func (t *MiTracer) Reset() {
	t.Lock()
	defer t.Unlock()
	t.finishedSpans = []*MiSpan{}
}

// StartSpan belongs to the Tracer interface.
func (t *MiTracer) StartSpan(operationName string, opts ...opentracing.StartSpanOption) opentracing.Span {
	sso := opentracing.StartSpanOptions{}
	for _, o := range opts {
		o.Apply(&sso)
	}
	return newMiSpan(t, operationName, sso)
}

// RegisterInjector registers injector for given format
func (t *MiTracer) RegisterInjector(format interface{}, injector Injector) {
	t.injectors[format] = injector
}

// RegisterExtractor registers extractor for given format
func (t *MiTracer) RegisterExtractor(format interface{}, extractor Extractor) {
	t.extractors[format] = extractor
}

// Inject belongs to the Tracer interface.
func (t *MiTracer) Inject(sm opentracing.SpanContext, format interface{}, carrier interface{}) error {
	spanContext, ok := sm.(MiSpanContext)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}
	injector, ok := t.injectors[format]
	if !ok {
		return opentracing.ErrUnsupportedFormat
	}
	return injector.Inject(spanContext, carrier)
}

// Extract belongs to the Tracer interface.
func (t *MiTracer) Extract(format interface{}, carrier interface{}) (opentracing.SpanContext, error) {
	extractor, ok := t.extractors[format]
	if !ok {
		return nil, opentracing.ErrUnsupportedFormat
	}
	return extractor.Extract(carrier)
}

func (t *MiTracer) recordSpan(span *MiSpan) {
	t.Lock()
	defer t.Unlock()
	t.finishedSpans = append(t.finishedSpans, span)
}
