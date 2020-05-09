package mitracer

import (
	"fmt"
	"reflect"
	"time"

	"github.com/opentracing/opentracing-go/log"
)

// MiLogRecord represents data logged to a Span via Span.LogFields or
// Span.LogKV.
type MiLogRecord struct {
	Timestamp time.Time
	Fields    []MiKeyValue
}

// MiKeyValue represents a single key:value pair.
type MiKeyValue struct {
	Key string

	// All MiLogRecord values are coerced to strings via fmt.Sprint(), though
	// we retain their type separately.
	ValueKind   reflect.Kind
	ValueString string
}

// EmitString belongs to the log.Encoder interface
func (m *MiKeyValue) EmitString(key, value string) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitBool belongs to the log.Encoder interface
func (m *MiKeyValue) EmitBool(key string, value bool) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitInt belongs to the log.Encoder interface
func (m *MiKeyValue) EmitInt(key string, value int) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitInt32 belongs to the log.Encoder interface
func (m *MiKeyValue) EmitInt32(key string, value int32) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitInt64 belongs to the log.Encoder interface
func (m *MiKeyValue) EmitInt64(key string, value int64) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitUint32 belongs to the log.Encoder interface
func (m *MiKeyValue) EmitUint32(key string, value uint32) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitUint64 belongs to the log.Encoder interface
func (m *MiKeyValue) EmitUint64(key string, value uint64) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitFloat32 belongs to the log.Encoder interface
func (m *MiKeyValue) EmitFloat32(key string, value float32) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitFloat64 belongs to the log.Encoder interface
func (m *MiKeyValue) EmitFloat64(key string, value float64) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitObject belongs to the log.Encoder interface
func (m *MiKeyValue) EmitObject(key string, value interface{}) {
	m.Key = key
	m.ValueKind = reflect.TypeOf(value).Kind()
	m.ValueString = fmt.Sprint(value)
}

// EmitLazyLogger belongs to the log.Encoder interface
func (m *MiKeyValue) EmitLazyLogger(value log.LazyLogger) {
	var meta MiKeyValue
	value(&meta)
	m.Key = meta.Key
	m.ValueKind = meta.ValueKind
	m.ValueString = meta.ValueString
}
