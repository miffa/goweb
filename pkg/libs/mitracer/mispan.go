package mitracer

import (
	"bytes"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// MiSpanContext is an opentracing.SpanContext implementation.
//
// It is entirely unsuitable for production use, but appropriate for tests
// that want to verify tracing behavior in other frameworks/applications.
//
// By default all spans have Sampled=true flag, unless {"sampling.priority": 0}
// tag is set.
type MiSpanContext struct {
	TraceID uint64
	SpanID  uint64
	Sampled bool
	Baggage map[string]string
}

var miIDSource = uint64(time.Now().Unix())

func nextMiID() uint64 {
	return atomic.AddUint64(&miIDSource, 1)
}

// ForeachBaggageItem belongs to the SpanContext interface
func (c MiSpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range c.Baggage {
		if !handler(k, v) {
			break
		}
	}
}

// WithBaggageItem creates a new context with an extra baggage item.
func (c MiSpanContext) WithBaggageItem(key, value string) MiSpanContext {
	var newBaggage map[string]string
	if c.Baggage == nil {
		newBaggage = map[string]string{key: value}
	} else {
		newBaggage = make(map[string]string, len(c.Baggage)+1)
		for k, v := range c.Baggage {
			newBaggage[k] = v
		}
		newBaggage[key] = value
	}
	// Use positional parameters so the compiler will help catch new fields.
	return MiSpanContext{c.TraceID, c.SpanID, c.Sampled, newBaggage}
}

// MiSpan is an opentracing.Span implementation that exports its internal
// state for testing purposes.
type MiSpan struct {
	sync.RWMutex

	ParentID uint64

	OperationName string
	StartTime     time.Time
	FinishTime    time.Time

	// All of the below are protected by the embedded RWMutex.
	SpanContext MiSpanContext
	tags        map[string]interface{}
	logs        []MiLogRecord
	tracer      *MiTracer
}

func newMiSpan(t *MiTracer, name string, opts opentracing.StartSpanOptions) *MiSpan {
	tags := opts.Tags
	if tags == nil {
		tags = map[string]interface{}{}
	}
	traceID := nextMiID()
	parentID := uint64(0)
	var baggage map[string]string
	sampled := true
	if len(opts.References) > 0 {
		traceID = opts.References[0].ReferencedContext.(MiSpanContext).TraceID
		parentID = opts.References[0].ReferencedContext.(MiSpanContext).SpanID
		sampled = opts.References[0].ReferencedContext.(MiSpanContext).Sampled
		baggage = opts.References[0].ReferencedContext.(MiSpanContext).Baggage
	}
	spanContext := MiSpanContext{traceID, nextMiID(), sampled, baggage}
	startTime := opts.StartTime
	if startTime.IsZero() {
		startTime = time.Now()
	}
	return &MiSpan{
		ParentID:      parentID,
		OperationName: name,
		StartTime:     startTime,
		tags:          tags,
		logs:          []MiLogRecord{},
		SpanContext:   spanContext,

		tracer: t,
	}
}

// Tags returns a copy of tags accumulated by the span so far
func (s *MiSpan) Tags() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()
	tags := make(map[string]interface{})
	for k, v := range s.tags {
		tags[k] = v
	}
	return tags
}

// Tag returns a single tag
func (s *MiSpan) Tag(k string) interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.tags[k]
}

// Logs returns a copy of logs accumulated in the span so far
func (s *MiSpan) Logs() []MiLogRecord {
	s.RLock()
	defer s.RUnlock()
	logs := make([]MiLogRecord, len(s.logs))
	copy(logs, s.logs)
	return logs
}

// Context belongs to the Span interface
func (s *MiSpan) Context() opentracing.SpanContext {
	s.Lock()
	defer s.Unlock()
	return s.SpanContext
}

// SetTag belongs to the Span interface
func (s *MiSpan) SetTag(key string, value interface{}) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	if key == string(ext.SamplingPriority) {
		if v, ok := value.(uint16); ok {
			s.SpanContext.Sampled = v > 0
			return s
		}
		if v, ok := value.(int); ok {
			s.SpanContext.Sampled = v > 0
			return s
		}
	}
	s.tags[key] = value
	return s
}

// SetBaggageItem belongs to the Span interface
func (s *MiSpan) SetBaggageItem(key, val string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.SpanContext = s.SpanContext.WithBaggageItem(key, val)
	return s
}

// BaggageItem belongs to the Span interface
func (s *MiSpan) BaggageItem(key string) string {
	s.RLock()
	defer s.RUnlock()
	return s.SpanContext.Baggage[key]
}

// Finish belongs to the Span interface
func (s *MiSpan) Finish() {
	s.Lock()
	s.FinishTime = time.Now()
	s.Unlock()
	s.tracer.recordSpan(s)
}

// FinishWithOptions belongs to the Span interface
func (s *MiSpan) FinishWithOptions(opts opentracing.FinishOptions) {
	s.Lock()
	s.FinishTime = opts.FinishTime
	s.Unlock()

	// Handle any late-bound LogRecords.
	for _, lr := range opts.LogRecords {
		s.logFieldsWithTimestamp(lr.Timestamp, lr.Fields...)
	}
	// Handle (deprecated) BulkLogData.
	for _, ld := range opts.BulkLogData {
		if ld.Payload != nil {
			s.logFieldsWithTimestamp(
				ld.Timestamp,
				log.String("event", ld.Event),
				log.Object("payload", ld.Payload))
		} else {
			s.logFieldsWithTimestamp(
				ld.Timestamp,
				log.String("event", ld.Event))
		}
	}

	s.tracer.recordSpan(s)
}

// String allows printing span for debugging
func (s *MiSpan) String() string {
	return s.Json()
	//return s.Text()
}

func (s *MiSpan) Text() string {
	return fmt.Sprintf(
		"traceId=%d, spanId=%d, parentId=%d, sampled=%t, name=%s, begin=%s, end=%s, cost=%d",
		s.SpanContext.TraceID,
		s.SpanContext.SpanID,
		s.ParentID,
		s.SpanContext.Sampled, s.OperationName,
		s.StartTime.String(),
		s.FinishTime.String(),
		(s.FinishTime.UnixNano() - s.StartTime.UnixNano()),
	)
}

func (s *MiSpan) Json() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("{")
	buf.WriteString("\"traceId\":")
	buf.WriteString(fmt.Sprintf("%d", s.SpanContext.TraceID))
	buf.WriteString(",\"spanId\":")
	buf.WriteString(fmt.Sprintf("%d", s.SpanContext.SpanID))
	buf.WriteString(",\"parentId\":")
	buf.WriteString(fmt.Sprintf("%d", s.ParentID))
	buf.WriteString(",\"startTime\":")
	buf.WriteString(fmt.Sprintf("%d", s.StartTime.UnixNano()/1000000))
	buf.WriteString(",\"finishTime\":")
	buf.WriteString(fmt.Sprintf("%d", s.FinishTime.UnixNano()/1000000))
	buf.WriteString(",\"costTime\":")
	buf.WriteString(fmt.Sprintf("%d", (s.FinishTime.UnixNano()-s.StartTime.UnixNano())/1000000))
	buf.WriteString(",\"sampled\":")
	buf.WriteString(fmt.Sprintf("%t", s.SpanContext.Sampled))
	buf.WriteString(",\"name\":\"")
	buf.WriteString(s.OperationName)
	buf.WriteString("\"")
	buf.WriteString("}")
	return buf.String()
}

// LogFields belongs to the Span interface
func (s *MiSpan) LogFields(fields ...log.Field) {
	s.logFieldsWithTimestamp(time.Now(), fields...)
}

// The caller MUST NOT hold s.Lock
func (s *MiSpan) logFieldsWithTimestamp(ts time.Time, fields ...log.Field) {
	lr := MiLogRecord{
		Timestamp: ts,
		Fields:    make([]MiKeyValue, len(fields)),
	}
	for i, f := range fields {
		outField := &(lr.Fields[i])
		f.Marshal(outField)
	}

	s.Lock()
	defer s.Unlock()
	s.logs = append(s.logs, lr)
}

// LogKV belongs to the Span interface.
//
// This implementations coerces all "values" to strings, though that is not
// something all implementations need to do. Indeed, a motivated person can and
// probably should have this do a typed switch on the values.
func (s *MiSpan) LogKV(keyValues ...interface{}) {
	if len(keyValues)%2 != 0 {
		s.LogFields(log.Error(fmt.Errorf("Non-even keyValues len: %v", len(keyValues))))
		return
	}
	fields, err := log.InterleavedKVToFields(keyValues...)
	if err != nil {
		s.LogFields(log.Error(err), log.String("function", "LogKV"))
		return
	}
	s.LogFields(fields...)
}

// LogEvent belongs to the Span interface
func (s *MiSpan) LogEvent(event string) {
	s.LogFields(log.String("event", event))
}

// LogEventWithPayload belongs to the Span interface
func (s *MiSpan) LogEventWithPayload(event string, payload interface{}) {
	s.LogFields(log.String("event", event), log.Object("payload", payload))
}

// Log belongs to the Span interface
func (s *MiSpan) Log(data opentracing.LogData) {
	panic("MiSpan.Log() no longer supported")
}

// SetOperationName belongs to the Span interface
func (s *MiSpan) SetOperationName(operationName string) opentracing.Span {
	s.Lock()
	defer s.Unlock()
	s.OperationName = operationName
	return s
}

// Tracer belongs to the Span interface
func (s *MiSpan) Tracer() opentracing.Tracer {
	return s.tracer
}

//func (s *MiSpan) MarshalJSON() ([]byte, error) {
//	buf := bytes.NewBuffer(nil)
//	buf.WriteString("{")
//
//	buf.WriteString("traceId:")
//	buf.WriteString(fmt.Sprintf("%d", s.SpanContext.TraceID))
//	buf.WriteString(",spanId:")
//	buf.WriteString(fmt.Sprintf("%d", s.SpanContext.SpanID))
//	buf.WriteString(",parentId:")
//	buf.WriteString(fmt.Sprintf("%d", s.SpanContext.ParentID))
//	buf.WriteString(",startTime:")
//	buf.WriteString(fmt.Sprintf("%d", s.StartTime.UnixNano()/1000000))
//	buf.WriteString(",finishTime:")
//	buf.WriteString(fmt.Sprintf("%d", s.FinishTime.UnixNano()/1000000))
//	buf.WriteString(", costTime:")
//	buf.WriteString(fmt.Sprintf("%d", (s.FinishTime.UnixNano() - s.StartTime.UnixNano()/1000000)))
//	buf.WriteString(",sampled:")
//	buf.WriteString(fmt.Sprintf("%t", s.SpanContext.Sampled))
//	buf.WriteString(",name:\"")
//	buf.WriteString(s.OperationName)
//	buf.WriteString("\"")
//	buf.WriteString("}")
//	return buf.Bytes()
//}
//
//func (s *MiSpan) UnmarshalJSON(data []byte) error {
//	if err := json.Unmarshal(data, s); err != nil {
//		return err
//	}
//	return nil
//}
