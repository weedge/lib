// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// from zap json_encoder.go change to encoder
package encoder

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/weedge/lib/log/internal/bufferpool"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// For JSON-escaping; see selfEncoder.safeAddString below.
const _hex = "0123456789abcdef"

var _selfPool = sync.Pool{New: func() interface{} {
	return &selfEncoder{}
}}

func getSelfEncoder() *selfEncoder {
	return _selfPool.Get().(*selfEncoder)
}

func putSelfEncoder(enc *selfEncoder) {
	if enc.reflectBuf != nil {
		enc.reflectBuf.Free()
	}
	enc.EncoderConfig = nil
	enc.buf = nil
	enc.spaced = false
	enc.openNamespaces = 0
	enc.reflectBuf = nil
	enc.reflectEnc = nil
	_selfPool.Put(enc)
}

type selfEncoder struct {
	*zapcore.EncoderConfig
	buf            *buffer.Buffer
	spaced         bool // include spaces after colons and commas
	openNamespaces int

	// for encoding generic values by reflection
	reflectBuf *buffer.Buffer
	reflectEnc *json.Encoder
}

// NewSelfEncoder creates a fast, low-allocation JSON encoder. The encoder
// appropriately escapes all field keys and values.
//
// Note that the encoder doesn't deduplicate keys, so it's possible to produce
// a message like
//   {"foo":"bar","foo":"baz"}
// This is permitted by the JSON specification, but not encouraged. Many
// libraries will ignore duplicate key-value pairs (typically keeping the last
// pair) when unmarshaling, but users should attempt to avoid adding duplicate
// keys.
func NewSelfEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return newSelfEncoder(cfg, false)
}

func newSelfEncoder(cfg zapcore.EncoderConfig, spaced bool) *selfEncoder {
	return &selfEncoder{
		EncoderConfig: &cfg,
		buf:           bufferpool.Get(),
		spaced:        spaced,
	}
}

func (enc *selfEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	err := enc.AppendArray(arr)
	enc.buf.AppendByte(']')
	return err
}

func (enc *selfEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	err := enc.AppendObject(obj)
	enc.buf.AppendByte(']')
	return err
}

func (enc *selfEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *selfEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendByteString(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendBool(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendComplex128(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendDuration(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendFloat64(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendInt64(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) resetReflectBuf() {
	if enc.reflectBuf == nil {
		enc.reflectBuf = bufferpool.Get()
		enc.reflectEnc = json.NewEncoder(enc.reflectBuf)
		// For consistency with our custom JSON encoder.
		enc.reflectEnc.SetEscapeHTML(false)
	} else {
		enc.reflectBuf.Reset()
	}
}

var nullLiteralBytes = []byte("null")

// Only invoke the standard JSON encoder if there is actually something to
// encode; otherwise write JSON null literal directly.
func (enc *selfEncoder) encodeReflected(obj interface{}) ([]byte, error) {
	if obj == nil {
		return nullLiteralBytes, nil
	}
	enc.resetReflectBuf()
	if err := enc.reflectEnc.Encode(obj); err != nil {
		return nil, err
	}
	enc.reflectBuf.TrimNewline()
	return enc.reflectBuf.Bytes(), nil
}

func (enc *selfEncoder) AddReflected(key string, obj interface{}) error {
	enc.addKey(key)
	valueBytes, err := enc.encodeReflected(obj)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *selfEncoder) OpenNamespace(key string) {
	enc.buf.AppendByte('{')
	enc.openNamespaces++
}

func (enc *selfEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendString(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendTime(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.buf.AppendByte('[')
	enc.AppendUint64(val)
	enc.buf.AppendByte(']')
}

func (enc *selfEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	enc.addElementSeparator()
	enc.buf.AppendByte('[')
	err := arr.MarshalLogArray(enc)
	enc.buf.AppendByte(']')
	return err
}

func (enc *selfEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	enc.addElementSeparator()
	enc.buf.AppendByte('{')
	err := obj.MarshalLogObject(enc)
	enc.buf.AppendByte('}')
	return err
}

func (enc *selfEncoder) AppendBool(val bool) {
	enc.addElementSeparator()
	enc.buf.AppendBool(val)
}

func (enc *selfEncoder) AppendByteString(val []byte) {
	enc.addElementSeparator()
	enc.buf.AppendByte('"')
	enc.safeAddByteString(val)
	enc.buf.AppendByte('"')
}

func (enc *selfEncoder) AppendComplex128(val complex128) {
	enc.addElementSeparator()
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendByte('"')
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, 64)
	enc.buf.AppendByte('+')
	enc.buf.AppendFloat(i, 64)
	enc.buf.AppendByte('i')
	enc.buf.AppendByte('"')
}

func (enc *selfEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	enc.EncodeDuration(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendInt64(int64(val))
	}
}

func (enc *selfEncoder) AppendInt64(val int64) {
	enc.addElementSeparator()
	enc.buf.AppendInt(val)
}

func (enc *selfEncoder) AppendReflected(val interface{}) error {
	valueBytes, err := enc.encodeReflected(val)
	if err != nil {
		return err
	}
	enc.addElementSeparator()
	_, err = enc.buf.Write(valueBytes)
	return err
}

func (enc *selfEncoder) AppendString(val string) {
	enc.addElementSeparator()
	enc.safeAddString(val)
}

func (enc *selfEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	enc.EncodeTime(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

func (enc *selfEncoder) AppendUint64(val uint64) {
	enc.addElementSeparator()
	enc.buf.AppendUint(val)
}

func (enc *selfEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }
func (enc *selfEncoder) AddFloat32(k string, v float32)     { enc.AddFloat64(k, float64(v)) }
func (enc *selfEncoder) AddInt(k string, v int)             { enc.AddInt64(k, int64(v)) }
func (enc *selfEncoder) AddInt32(k string, v int32)         { enc.AddInt64(k, int64(v)) }
func (enc *selfEncoder) AddInt16(k string, v int16)         { enc.AddInt64(k, int64(v)) }
func (enc *selfEncoder) AddInt8(k string, v int8)           { enc.AddInt64(k, int64(v)) }
func (enc *selfEncoder) AddUint(k string, v uint)           { enc.AddUint64(k, uint64(v)) }
func (enc *selfEncoder) AddUint32(k string, v uint32)       { enc.AddUint64(k, uint64(v)) }
func (enc *selfEncoder) AddUint16(k string, v uint16)       { enc.AddUint64(k, uint64(v)) }
func (enc *selfEncoder) AddUint8(k string, v uint8)         { enc.AddUint64(k, uint64(v)) }
func (enc *selfEncoder) AddUintptr(k string, v uintptr)     { enc.AddUint64(k, uint64(v)) }
func (enc *selfEncoder) AppendComplex64(v complex64)        { enc.AppendComplex128(complex128(v)) }
func (enc *selfEncoder) AppendFloat64(v float64)            { enc.appendFloat(v, 64) }
func (enc *selfEncoder) AppendFloat32(v float32)            { enc.appendFloat(float64(v), 32) }
func (enc *selfEncoder) AppendInt(v int)                    { enc.AppendInt64(int64(v)) }
func (enc *selfEncoder) AppendInt32(v int32)                { enc.AppendInt64(int64(v)) }
func (enc *selfEncoder) AppendInt16(v int16)                { enc.AppendInt64(int64(v)) }
func (enc *selfEncoder) AppendInt8(v int8)                  { enc.AppendInt64(int64(v)) }
func (enc *selfEncoder) AppendUint(v uint)                  { enc.AppendUint64(uint64(v)) }
func (enc *selfEncoder) AppendUint32(v uint32)              { enc.AppendUint64(uint64(v)) }
func (enc *selfEncoder) AppendUint16(v uint16)              { enc.AppendUint64(uint64(v)) }
func (enc *selfEncoder) AppendUint8(v uint8)                { enc.AppendUint64(uint64(v)) }
func (enc *selfEncoder) AppendUintptr(v uintptr)            { enc.AppendUint64(uint64(v)) }

func (enc *selfEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	clone.buf.Write(enc.buf.Bytes())
	return clone
}

func (enc *selfEncoder) clone() *selfEncoder {
	clone := getSelfEncoder()
	clone.EncoderConfig = enc.EncoderConfig
	clone.spaced = enc.spaced
	clone.openNamespaces = enc.openNamespaces
	clone.buf = bufferpool.Get()
	return clone
}

func (enc *selfEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()

	if final.LevelKey != "" {
		final.buf.AppendByte('[')
		//final.addKey(final.LevelKey)
		cur := final.buf.Len()
		final.EncodeLevel(ent.Level, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeLevel was a no-op. Fall back to strings to keep
			// output JSON valid.
			final.AppendString(ent.Level.String())
		}
		final.buf.AppendByte(']')
	}
	if final.TimeKey != "" {
		final.buf.AppendByte(' ')
		final.AppendTime(ent.Time)
	}
	if ent.LoggerName != "" && final.NameKey != "" {
		final.buf.AppendByte(' ')
		cur := final.buf.Len()
		nameEncoder := final.EncodeName

		// if no name encoder provided, fall back to FullNameEncoder for backwards
		// compatibility
		if nameEncoder == nil {
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(ent.LoggerName, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeName was a no-op. Fall back to strings to
			// keep output JSON valid.
			final.AppendString(ent.LoggerName)
		}
	}
	if ent.Caller.Defined && final.CallerKey != "" {
		final.buf.AppendByte(' ')
		final.buf.AppendByte('[')
		cur := final.buf.Len()
		final.EncodeCaller(ent.Caller, final)
		final.buf.AppendByte(']')
		if cur == final.buf.Len() {
			// User-supplied EncodeCaller was a no-op. Fall back to strings to
			// keep output JSON valid.
			final.AppendString(ent.Caller.String())
		}
	} else {
		final.buf.AppendString(" [access]")
	}
	if final.MessageKey != "" {
		final.AppendString(ent.Message)
	}
	if enc.buf.Len() > 0 {
		final.addElementSeparator()
		final.buf.Write(enc.buf.Bytes())
	}
	addFields(final, fields)
	final.closeOpenNamespaces()
	if ent.Stack != "" && final.StacktraceKey != "" {
		final.AddString(final.StacktraceKey, ent.Stack)
	}
	if final.LineEnding != "" {
		final.buf.AppendString(final.LineEnding)
	} else {
		final.buf.AppendString(zapcore.DefaultLineEnding)
	}

	ret := final.buf
	putSelfEncoder(final)
	return ret, nil
}

func (enc *selfEncoder) truncate() {
	enc.buf.Reset()
}

func (enc *selfEncoder) closeOpenNamespaces() {
	for i := 0; i < enc.openNamespaces; i++ {
		enc.buf.AppendByte('}')
	}
}

func (enc *selfEncoder) addKey(key string) {
	enc.addElementSeparator()
	enc.safeAddString(key)
}

func (enc *selfEncoder) addElementSeparator() {
	last := enc.buf.Len() - 1
	if last < 0 {
		return
	}
	switch enc.buf.Bytes()[last] {
	case '{', '[', ':', ',', ' ':
		return
	default:
		enc.buf.AppendByte(' ')
	}
}

func (enc *selfEncoder) appendFloat(val float64, bitSize int) {
	enc.addElementSeparator()
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *selfEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

// safeAddByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (enc *selfEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.Write(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (enc *selfEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		// enc.buf.AppendByte('\\')
		enc.buf.AppendByte(b)
	case '\n':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('n')
	case '\r':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('r')
	case '\t':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (enc *selfEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}
