package cborwire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// Decode options.
type DecodeOptions struct {
	// MaxBytes limits total input size (0 = unlimited).
	MaxBytes int
	// MaxDepth limits nesting of arrays/maps/tags (0 = default 32).
	MaxDepth int
	// DecodeTags expands well-known tags (0,1,2,3) into richer Values.
	DecodeTags bool
}

// DefaultDecodeOptions returns sensible defaults.
func DefaultDecodeOptions() DecodeOptions {
	return DecodeOptions{MaxDepth: 32, DecodeTags: true}
}

// Decode decodes one CBOR data item from b using default options.
func Decode(b []byte) (Value, error) {
	return DecodeWith(b, DefaultDecodeOptions())
}

// DecodeWith decodes one CBOR data item with options.
func DecodeWith(b []byte, opt DecodeOptions) (Value, error) {
	if opt.MaxDepth <= 0 {
		opt.MaxDepth = 32
	}
	if opt.MaxBytes > 0 && len(b) > opt.MaxBytes {
		return Value{}, fmt.Errorf("cborwire: input exceeds MaxBytes %d", opt.MaxBytes)
	}
	d := &decoder{buf: b, opt: opt}
	v, err := d.decodeItem(0)
	if err != nil {
		return Value{}, err
	}
	return v, nil
}

// DecodeAll decodes consecutive CBOR items until EOF.
func DecodeAll(b []byte, opt DecodeOptions) ([]Value, error) {
	if opt.MaxDepth <= 0 {
		opt.MaxDepth = 32
	}
	d := &decoder{buf: b, opt: opt}
	var out []Value
	for d.off < len(d.buf) {
		v, err := d.decodeItem(0)
		if err != nil {
			return out, err
		}
		out = append(out, v)
	}
	return out, nil
}

type decoder struct {
	buf []byte
	off int
	opt DecodeOptions
}

func (d *decoder) remaining() int { return len(d.buf) - d.off }

func (d *decoder) readByte() (byte, error) {
	if d.off >= len(d.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	b := d.buf[d.off]
	d.off++
	return b, nil
}

func (d *decoder) readN(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("cborwire: negative length")
	}
	if d.remaining() < n {
		return nil, io.ErrUnexpectedEOF
	}
	b := d.buf[d.off : d.off+n]
	d.off += n
	return b, nil
}

func (d *decoder) decodeItem(depth int) (Value, error) {
	if depth > d.opt.MaxDepth {
		return Value{}, fmt.Errorf("cborwire: max depth %d exceeded", d.opt.MaxDepth)
	}
	ib, err := d.readByte()
	if err != nil {
		return Value{}, err
	}
	mt := MajorType(ib >> 5)
	ai := ib & 0x1f

	switch mt {
	case MajorUnsigned:
		n, err := d.readArgument(ai)
		if err != nil {
			return Value{}, err
		}
		return Uint(n), nil

	case MajorNegative:
		n, err := d.readArgument(ai)
		if err != nil {
			return Value{}, err
		}
		return NegInt(n), nil

	case MajorBytes:
		if ai == 31 {
			return Value{}, errors.New("cborwire: indefinite byte string not supported in this stage")
		}
		n, err := d.readArgument(ai)
		if err != nil {
			return Value{}, err
		}
		if n > uint64(d.remaining()) {
			return Value{}, io.ErrUnexpectedEOF
		}
		b, err := d.readN(int(n))
		if err != nil {
			return Value{}, err
		}
		return Bytes(b), nil

	case MajorText:
		if ai == 31 {
			return Value{}, errors.New("cborwire: indefinite text string not supported in this stage")
		}
		n, err := d.readArgument(ai)
		if err != nil {
			return Value{}, err
		}
		if n > uint64(d.remaining()) {
			return Value{}, io.ErrUnexpectedEOF
		}
		b, err := d.readN(int(n))
		if err != nil {
			return Value{}, err
		}
		if !validUTF8(b) {
			return Value{}, errors.New("cborwire: text string is not valid UTF-8")
		}
		return Text(string(b)), nil

	case MajorArray:
		if ai == 31 {
			return Value{}, errors.New("cborwire: indefinite array not supported in this stage")
		}
		n, err := d.readArgument(ai)
		if err != nil {
			return Value{}, err
		}
		if n > uint64(d.remaining()) { // rough bound
			return Value{}, io.ErrUnexpectedEOF
		}
		arr := make([]Value, 0, int(min64(n, 64)))
		for i := uint64(0); i < n; i++ {
			v, err := d.decodeItem(depth + 1)
			if err != nil {
				return Value{}, err
			}
			arr = append(arr, v)
		}
		return Array(arr...), nil

	case MajorMap:
		if ai == 31 {
			return Value{}, errors.New("cborwire: indefinite map not supported in this stage")
		}
		n, err := d.readArgument(ai)
		if err != nil {
			return Value{}, err
		}
		pairs := make([]Pair, 0, int(min64(n, 64)))
		for i := uint64(0); i < n; i++ {
			k, err := d.decodeItem(depth + 1)
			if err != nil {
				return Value{}, err
			}
			v, err := d.decodeItem(depth + 1)
			if err != nil {
				return Value{}, err
			}
			pairs = append(pairs, Pair{Key: k, Value: v})
		}
		return MapOf(pairs...), nil

	case MajorTag:
		n, err := d.readArgument(ai)
		if err != nil {
			return Value{}, err
		}
		inner, err := d.decodeItem(depth + 1)
		if err != nil {
			return Value{}, err
		}
		if d.opt.DecodeTags {
			if v, ok := expandTag(n, inner); ok {
				return v, nil
			}
		}
		return Tag(n, inner), nil

	case MajorSimple:
		return d.decodeSimple(ai)

	default:
		return Value{}, fmt.Errorf("cborwire: unknown major type %d", mt)
	}
}

func (d *decoder) decodeSimple(ai byte) (Value, error) {
	switch ai {
	case 20:
		return Bool(false), nil
	case 21:
		return Bool(true), nil
	case 22:
		return Null(), nil
	case 23:
		return Undefined(), nil
	case 24:
		b, err := d.readByte()
		if err != nil {
			return Value{}, err
		}
		if b < 32 {
			return Value{}, fmt.Errorf("cborwire: invalid simple value %d", b)
		}
		return Simple(b), nil
	case 25:
		raw, err := d.readN(2)
		if err != nil {
			return Value{}, err
		}
		u := binary.BigEndian.Uint16(raw)
		return Float(float64(math.Float32frombits(float16ToFloat32(u)))), nil
	case 26:
		raw, err := d.readN(4)
		if err != nil {
			return Value{}, err
		}
		return Float(float64(math.Float32frombits(binary.BigEndian.Uint32(raw)))), nil
	case 27:
		raw, err := d.readN(8)
		if err != nil {
			return Value{}, err
		}
		return Float(math.Float64frombits(binary.BigEndian.Uint64(raw))), nil
	case 31:
		return Value{}, errors.New("cborwire: unexpected break code")
	default:
		if ai < 20 {
			return Simple(ai), nil
		}
		return Value{}, fmt.Errorf("cborwire: reserved simple ai=%d", ai)
	}
}

func (d *decoder) readArgument(ai byte) (uint64, error) {
	switch {
	case ai < 24:
		return uint64(ai), nil
	case ai == 24:
		b, err := d.readByte()
		if err != nil {
			return 0, err
		}
		return uint64(b), nil
	case ai == 25:
		raw, err := d.readN(2)
		if err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint16(raw)), nil
	case ai == 26:
		raw, err := d.readN(4)
		if err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint32(raw)), nil
	case ai == 27:
		raw, err := d.readN(8)
		if err != nil {
			return 0, err
		}
		return binary.BigEndian.Uint64(raw), nil
	default:
		return 0, fmt.Errorf("cborwire: reserved additional info %d", ai)
	}
}

func min64(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func validUTF8(b []byte) bool {
	// stdlib unicode/utf8
	for i := 0; i < len(b); {
		if b[i] < 0x80 {
			i++
			continue
		}
		// use utf8.DecodeRune
		r, size := decodeRune(b[i:])
		if r == 0xfffd && size == 1 {
			return false
		}
		i += size
	}
	return true
}

// Minimal UTF-8 decode to avoid importing in every call path description;
// delegates to unicode/utf8 via a thin wrapper in utf8.go.
func decodeRune(p []byte) (rune, int) {
	return decodeRuneImpl(p)
}

// expandTag is a stub in stage 1; tags 0-3 fully handled in stage 2.
func expandTag(num uint64, inner Value) (Value, bool) {
	_ = num
	_ = inner
	return Value{}, false
}

// float16ToFloat32 is a stub; full half-float in stage 2.
func float16ToFloat32(u uint16) uint32 {
	// IEEE 754 half → single (correct conversion for stage-1 float tests that use 0/inf/nan carefully).
	sign := uint32(u&0x8000) << 16
	exp := (u >> 10) & 0x1f
	frac := uint32(u & 0x3ff)
	switch exp {
	case 0:
		if frac == 0 {
			return sign
		}
		// subnormal
		exp32 := uint32(127 - 15 + 1)
		for frac&0x400 == 0 {
			frac <<= 1
			exp32--
		}
		frac &= 0x3ff
		return sign | (exp32 << 23) | (frac << 13)
	case 31:
		if frac == 0 {
			return sign | 0x7f800000
		}
		return sign | 0x7f800000 | (frac << 13)
	default:
		return sign | (uint32(exp-15+127) << 23) | (frac << 13)
	}
}
