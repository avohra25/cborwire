package cborwire

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/big"
)

// EncodeOptions controls serialization.
type EncodeOptions struct {
	// Canonical uses preferred/shortest integer encodings and (in later
	// stages) sorted map keys per RFC 8949 §4.2.
	Canonical bool
	// PreferredFloat prefers float16/float32 when they round-trip exactly.
	PreferredFloat bool
}

// Encode encodes v to CBOR bytes with default (non-canonical) options.
func Encode(v Value) ([]byte, error) {
	return EncodeWith(v, EncodeOptions{})
}

// EncodeWith encodes v with options.
func EncodeWith(v Value, opt EncodeOptions) ([]byte, error) {
	var buf []byte
	w := &byteWriter{buf: &buf}
	var err error
	if opt.Canonical {
		err = encodeValueCanonical(w, v, opt)
	} else {
		err = encodeValue(w, v, opt)
	}
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// MustEncode panics on error; useful in tests.
func MustEncode(v Value) []byte {
	b, err := Encode(v)
	if err != nil {
		panic(err)
	}
	return b
}

type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func (w *byteWriter) WriteByte(b byte) error {
	*w.buf = append(*w.buf, b)
	return nil
}

type writer interface {
	Write([]byte) (int, error)
	WriteByte(byte) error
}

func encodeValue(w writer, v Value, opt EncodeOptions) error {
	switch v.Kind {
	case KindUnsigned:
		return writeTypeArg(w, MajorUnsigned, v.Uint)
	case KindNegative:
		return writeTypeArg(w, MajorNegative, v.Neg)
	case KindBytes:
		if err := writeTypeArg(w, MajorBytes, uint64(len(v.Bytes))); err != nil {
			return err
		}
		_, err := w.Write(v.Bytes)
		return err
	case KindText:
		b := []byte(v.Text)
		if err := writeTypeArg(w, MajorText, uint64(len(b))); err != nil {
			return err
		}
		_, err := w.Write(b)
		return err
	case KindArray:
		if err := writeTypeArg(w, MajorArray, uint64(len(v.Array))); err != nil {
			return err
		}
		for _, e := range v.Array {
			if err := encodeValue(w, e, opt); err != nil {
				return err
			}
		}
		return nil
	case KindMap:
		pairs := v.Map
		if err := writeTypeArg(w, MajorMap, uint64(len(pairs))); err != nil {
			return err
		}
		for _, p := range pairs {
			if err := encodeValue(w, p.Key, opt); err != nil {
				return err
			}
			if err := encodeValue(w, p.Value, opt); err != nil {
				return err
			}
		}
		return nil
	case KindTag:
		if v.TagVal == nil {
			return fmt.Errorf("cborwire: tag %d has nil content", v.TagNum)
		}
		if err := writeTypeArg(w, MajorTag, v.TagNum); err != nil {
			return err
		}
		return encodeValue(w, *v.TagVal, opt)
	case KindBool:
		if v.Bool {
			return w.WriteByte(0xf5)
		}
		return w.WriteByte(0xf4)
	case KindNull:
		return w.WriteByte(0xf6)
	case KindUndefined:
		return w.WriteByte(0xf7)
	case KindFloat:
		return encodeFloat(w, v.Float, opt)
	case KindSimple:
		if v.Simple < 24 {
			return w.WriteByte(0xe0 | v.Simple)
		}
		if err := w.WriteByte(0xf8); err != nil {
			return err
		}
		return w.WriteByte(v.Simple)
	case KindBigInt:
		return encodeBigInt(w, v.BigInt, opt)
	default:
		return fmt.Errorf("cborwire: cannot encode Kind %d", v.Kind)
	}
}

func encodeBigInt(w writer, n *big.Int, opt EncodeOptions) error {
	if n == nil {
		return fmt.Errorf("cborwire: nil big.Int")
	}
	if n.Sign() >= 0 {
		// Prefer major type 0 when it fits in uint64.
		if n.IsUint64() {
			return writeTypeArg(w, MajorUnsigned, n.Uint64())
		}
		b := n.Bytes()
		if err := writeTypeArg(w, MajorTag, 2); err != nil {
			return err
		}
		return encodeValue(w, Bytes(b), opt)
	}
	// negative: tag 3 with bytes of (-1 - n) = -(n+1)
	abs := new(big.Int).Neg(n)
	abs.Sub(abs, big.NewInt(1))
	if abs.IsUint64() {
		return writeTypeArg(w, MajorNegative, abs.Uint64())
	}
	b := abs.Bytes()
	if err := writeTypeArg(w, MajorTag, 3); err != nil {
		return err
	}
	return encodeValue(w, Bytes(b), opt)
}

func encodeFloat(w writer, f float64, opt EncodeOptions) error {
	if opt.PreferredFloat || opt.Canonical {
		// try float16
		if u, ok := float64ToFloat16(f); ok {
			if err := w.WriteByte(0xf9); err != nil {
				return err
			}
			var tmp [2]byte
			binary.BigEndian.PutUint16(tmp[:], u)
			_, err := w.Write(tmp[:])
			return err
		}
		// try float32
		f32 := float32(f)
		if float64(f32) == f || math.IsNaN(f) {
			if err := w.WriteByte(0xfa); err != nil {
				return err
			}
			var tmp [4]byte
			binary.BigEndian.PutUint32(tmp[:], math.Float32bits(f32))
			_, err := w.Write(tmp[:])
			return err
		}
	}
	if err := w.WriteByte(0xfb); err != nil {
		return err
	}
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], math.Float64bits(f))
	_, err := w.Write(tmp[:])
	return err
}

func writeTypeArg(w writer, mt MajorType, n uint64) error {
	head := byte(mt) << 5
	switch {
	case n < 24:
		return w.WriteByte(head | byte(n))
	case n <= 0xff:
		if err := w.WriteByte(head | 24); err != nil {
			return err
		}
		return w.WriteByte(byte(n))
	case n <= 0xffff:
		if err := w.WriteByte(head | 25); err != nil {
			return err
		}
		var tmp [2]byte
		binary.BigEndian.PutUint16(tmp[:], uint16(n))
		_, err := w.Write(tmp[:])
		return err
	case n <= 0xffffffff:
		if err := w.WriteByte(head | 26); err != nil {
			return err
		}
		var tmp [4]byte
		binary.BigEndian.PutUint32(tmp[:], uint32(n))
		_, err := w.Write(tmp[:])
		return err
	default:
		if err := w.WriteByte(head | 27); err != nil {
			return err
		}
		var tmp [8]byte
		binary.BigEndian.PutUint64(tmp[:], n)
		_, err := w.Write(tmp[:])
		return err
	}
}

// float64ToFloat16 tries exact round-trip conversion; returns ok=false if not exact.
func float64ToFloat16(f float64) (uint16, bool) {
	if math.IsNaN(f) {
		return 0x7e00, true // canonical qNaN
	}
	if math.IsInf(f, 1) {
		return 0x7c00, true
	}
	if math.IsInf(f, -1) {
		return 0xfc00, true
	}
	bits := math.Float64bits(f)
	sign := uint16((bits >> 63) << 15)
	exp := int((bits>>52)&0x7ff) - 1023
	frac := bits & 0xfffffffffffff

	if exp < -14 {
		// subnormal or zero in float16
		if exp < -24 {
			return sign, f == 0 || (sign != 0 && f == 0) // +0/-0
		}
		// try subnormal
		shift := -14 - exp + 10 + (52 - 10) // complicated; use float32 path
		_ = shift
		f32 := float32(f)
		u := float32ToFloat16Bits(f32)
		back := math.Float32frombits(float16ToFloat32(u))
		if float64(back) == f {
			return u, true
		}
		return 0, false
	}
	if exp > 15 {
		return 0, false
	}
	// normal
	if frac&((1<<(52-10))-1) != 0 {
		// would need rounding; check exactness via round-trip
		f32 := float32(f)
		u := float32ToFloat16Bits(f32)
		back := math.Float32frombits(float16ToFloat32(u))
		if float64(back) == f {
			return u, true
		}
		return 0, false
	}
	mant := uint16(frac >> (52 - 10))
	return sign | uint16(exp+15)<<10 | mant, true
}

func float32ToFloat16Bits(f float32) uint16 {
	bits := math.Float32bits(f)
	sign := uint16((bits >> 16) & 0x8000)
	exp := int((bits>>23)&0xff) - 127 + 15
	frac := bits & 0x7fffff
	switch {
	case exp <= 0:
		if exp < -10 {
			return sign
		}
		frac |= 0x800000
		frac >>= uint(1 - exp)
		return sign | uint16(frac>>13)
	case exp >= 31:
		if frac == 0 {
			return sign | 0x7c00
		}
		return sign | 0x7e00
	default:
		return sign | uint16(exp)<<10 | uint16(frac>>13)
	}
}

// EncodeTo writes CBOR to an io.Writer (stage-1 convenience; full streaming in stage 3).
func EncodeTo(w io.Writer, v Value) error {
	b, err := Encode(v)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}
