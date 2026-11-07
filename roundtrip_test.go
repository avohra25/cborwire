package cborwire

import (
	"bytes"
	"math"
	"testing"
)

func TestRoundTripValues(t *testing.T) {
	vals := []Value{
		Uint(0), Uint(23), Uint(24), Uint(255), Uint(256), Uint(65535), Uint(65536),
		Uint(1 << 32), Uint(^uint64(0)),
		NegInt(0), NegInt(23), NegInt(24), NegInt(255), NegInt(1000),
		Bytes(nil), Bytes([]byte{}), Bytes([]byte{0}), Bytes([]byte{1, 2, 3, 4, 5}),
		Text(""), Text("a"), Text("hello"), Text("你好"), Text("𝄞"),
		Array(), Array(Uint(1)), Array(Uint(1), Text("x"), Null()),
		Array(Array(Uint(1), Uint(2)), Array(Uint(3))),
		MapOf(), MapOf(Pair{Key: Uint(1), Value: Uint(2)}),
		MapOf(Pair{Key: Text("a"), Value: Array(Uint(1), Bool(true))}),
		Bool(false), Bool(true), Null(), Undefined(),
		Float(0), Float(1.5), Float(-2.5), Float(math.Pi),
		Float(math.Inf(1)), Float(math.Inf(-1)),
		Tag(0, Text("2013-03-21T20:04:00Z")),
		Tag(1, Uint(1363896240)),
		Tag(23, Bytes([]byte{1, 2, 3})),
		Simple(32), Simple(255),
		FromInt(-9223372036854775807),
		FromInt(9223372036854775807),
	}
	for i, v := range vals {
		b, err := Encode(v)
		if err != nil {
			t.Fatalf("%d encode: %v", i, err)
		}
		got, err := Decode(b)
		if err != nil {
			t.Fatalf("%d decode: %v", i, err)
		}
		b2, err := Encode(got)
		if err != nil {
			t.Fatalf("%d re-encode: %v", i, err)
		}
		if !bytes.Equal(b, b2) {
			t.Fatalf("%d: encode not stable: %x vs %x", i, b, b2)
		}
	}
}

func TestRoundTripNaN(t *testing.T) {
	v := Float(math.NaN())
	b, err := Encode(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if !math.IsNaN(got.Float) {
		t.Fatalf("got %v", got.Float)
	}
}

func TestPreferredFloatRoundTrip(t *testing.T) {
	opt := EncodeOptions{PreferredFloat: true}
	for _, f := range []float64{0, 1, 1.5, -2, 65504} {
		b, err := EncodeWith(Float(f), opt)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Decode(b)
		if err != nil {
			t.Fatal(err)
		}
		if got.Float != f {
			t.Fatalf("%v -> %v (bytes %x)", f, got.Float, b)
		}
	}
}
