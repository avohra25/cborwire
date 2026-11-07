// Package cborwire is a pure-Go CBOR (RFC 8949) codec toolkit.
package cborwire

import (
	"fmt"
	"math/big"
)

// MajorType is a CBOR major type (0–7).
type MajorType byte

const (
	MajorUnsigned MajorType = 0
	MajorNegative MajorType = 1
	MajorBytes    MajorType = 2
	MajorText     MajorType = 3
	MajorArray    MajorType = 4
	MajorMap      MajorType = 5
	MajorTag      MajorType = 6
	MajorSimple   MajorType = 7
)

// Kind identifies the decoded value kind after tags are expanded.
type Kind int

const (
	KindUnsigned Kind = iota
	KindNegative
	KindBytes
	KindText
	KindArray
	KindMap
	KindTag
	KindBool
	KindNull
	KindUndefined
	KindFloat
	KindSimple
	KindBigInt // tags 2/3 decoded to *big.Int when DecodeTags is set
)

// Pair is a map entry preserving CBOR key order.
type Pair struct {
	Key   Value
	Value Value
}

// Value is a decoded CBOR data item.
type Value struct {
	Kind Kind

	Uint    uint64
	Neg     uint64   // magnitude of negative int: -1-Neg
	Bytes   []byte
	Text    string
	Array   []Value
	Map     []Pair
	TagNum  uint64
	TagVal  *Value
	Bool    bool
	Float   float64
	Simple  byte
	BigInt  *big.Int
	RawTag  bool // true when Tag was left wrapped (unknown / DecodeTags off)
}

// IsNull reports whether v is the CBOR null simple value.
func (v Value) IsNull() bool { return v.Kind == KindNull }

// IsUndefined reports whether v is the CBOR undefined simple value.
func (v Value) IsUndefined() bool { return v.Kind == KindUndefined }

// Int64 converts unsigned/negative integers that fit in int64.
func (v Value) Int64() (int64, error) {
	switch v.Kind {
	case KindUnsigned:
		if v.Uint > uint64(^uint64(0)>>1) {
			return 0, fmt.Errorf("cborwire: unsigned %d overflows int64", v.Uint)
		}
		return int64(v.Uint), nil
	case KindNegative:
		// value = -1 - Neg
		if v.Neg >= uint64(^uint64(0)>>1) {
			return 0, fmt.Errorf("cborwire: negative -1-%d overflows int64", v.Neg)
		}
		return -1 - int64(v.Neg), nil
	case KindBigInt:
		if v.BigInt == nil || !v.BigInt.IsInt64() {
			return 0, fmt.Errorf("cborwire: big int does not fit in int64")
		}
		return v.BigInt.Int64(), nil
	default:
		return 0, fmt.Errorf("cborwire: Kind %d is not an integer", v.Kind)
	}
}

// Uint64 returns the unsigned integer value.
func (v Value) Uint64() (uint64, error) {
	if v.Kind != KindUnsigned {
		return 0, fmt.Errorf("cborwire: Kind %d is not unsigned", v.Kind)
	}
	return v.Uint, nil
}

// AsBigInt returns a *big.Int for integer kinds.
func (v Value) AsBigInt() (*big.Int, error) {
	switch v.Kind {
	case KindUnsigned:
		return new(big.Int).SetUint64(v.Uint), nil
	case KindNegative:
		n := new(big.Int).SetUint64(v.Neg)
		n.Add(n, big.NewInt(1))
		n.Neg(n)
		return n, nil
	case KindBigInt:
		if v.BigInt == nil {
			return nil, fmt.Errorf("cborwire: nil big int")
		}
		return new(big.Int).Set(v.BigInt), nil
	default:
		return nil, fmt.Errorf("cborwire: Kind %d is not an integer", v.Kind)
	}
}

// Clone deep-copies a Value.
func (v Value) Clone() Value {
	out := v
	if v.Bytes != nil {
		out.Bytes = append([]byte(nil), v.Bytes...)
	}
	if v.Array != nil {
		out.Array = make([]Value, len(v.Array))
		for i := range v.Array {
			out.Array[i] = v.Array[i].Clone()
		}
	}
	if v.Map != nil {
		out.Map = make([]Pair, len(v.Map))
		for i := range v.Map {
			out.Map[i].Key = v.Map[i].Key.Clone()
			out.Map[i].Value = v.Map[i].Value.Clone()
		}
	}
	if v.TagVal != nil {
		tv := v.TagVal.Clone()
		out.TagVal = &tv
	}
	if v.BigInt != nil {
		out.BigInt = new(big.Int).Set(v.BigInt)
	}
	return out
}

// Helpers to construct Values.

func Uint(n uint64) Value       { return Value{Kind: KindUnsigned, Uint: n} }
func NegInt(mag uint64) Value   { return Value{Kind: KindNegative, Neg: mag} }
func Bytes(b []byte) Value      { return Value{Kind: KindBytes, Bytes: append([]byte(nil), b...)} }
func Text(s string) Value       { return Value{Kind: KindText, Text: s} }
func Array(vs ...Value) Value   { return Value{Kind: KindArray, Array: vs} }
func Bool(b bool) Value         { return Value{Kind: KindBool, Bool: b} }
func Null() Value               { return Value{Kind: KindNull} }
func Undefined() Value          { return Value{Kind: KindUndefined} }
func Float(f float64) Value     { return Value{Kind: KindFloat, Float: f} }
func Simple(s byte) Value       { return Value{Kind: KindSimple, Simple: s} }
func Tag(n uint64, v Value) Value {
	vv := v.Clone()
	return Value{Kind: KindTag, TagNum: n, TagVal: &vv, RawTag: true}
}
func BigInt(n *big.Int) Value {
	return Value{Kind: KindBigInt, BigInt: new(big.Int).Set(n)}
}

// MapOf builds a map from alternating key/value pairs.
func MapOf(pairs ...Pair) Value {
	return Value{Kind: KindMap, Map: pairs}
}

// FromInt converts a Go int64 to a CBOR integer Value.
func FromInt(n int64) Value {
	if n >= 0 {
		return Uint(uint64(n))
	}
	return NegInt(uint64(-n) - 1)
}
