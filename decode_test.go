package cborwire

import (
	"bytes"
	"encoding/hex"
	"math"
	"testing"
)

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestRFCAppendixA_Integers(t *testing.T) {
	cases := []struct {
		hex string
		n   int64
	}{
		{"00", 0},
		{"01", 1},
		{"0a", 10},
		{"17", 23},
		{"1818", 24},
		{"1819", 25},
		{"1864", 100},
		{"1903e8", 1000},
		{"1a000f4240", 1000000},
		{"1b000000e8d4a51000", 1000000000000},
		{"20", -1},
		{"29", -10},
		{"3863", -100},
		{"3903e7", -1000},
	}
	for _, c := range cases {
		v, err := Decode(mustHex(c.hex))
		if err != nil {
			t.Fatalf("%s: decode: %v", c.hex, err)
		}
		got, err := v.Int64()
		if err != nil {
			t.Fatalf("%s: Int64: %v", c.hex, err)
		}
		if got != c.n {
			t.Fatalf("%s: got %d want %d", c.hex, got, c.n)
		}
		enc, err := Encode(v)
		if err != nil {
			t.Fatalf("%s: encode: %v", c.hex, err)
		}
		if !bytes.Equal(enc, mustHex(c.hex)) {
			t.Fatalf("%s: round-trip encode got %x", c.hex, enc)
		}
	}
}

func TestRFCAppendixA_BytesText(t *testing.T) {
	cases := []struct {
		hex  string
		kind Kind
		text string
		raw  string
	}{
		{"40", KindBytes, "", ""},
		{"4401020304", KindBytes, "", "\x01\x02\x03\x04"},
		{"60", KindText, "", ""},
		{"6161", KindText, "a", ""},
		{"6449455446", KindText, "IETF", ""},
		{"62225c", KindText, "\"\\", ""},
		{"62c3bc", KindText, "ü", ""},
		{"63e6b0b4", KindText, "水", ""},
		{"64f0908591", KindText, "𐅑", ""},
	}
	for _, c := range cases {
		v, err := Decode(mustHex(c.hex))
		if err != nil {
			t.Fatalf("%s: %v", c.hex, err)
		}
		if v.Kind != c.kind {
			t.Fatalf("%s: kind %v", c.hex, v.Kind)
		}
		if c.kind == KindText && v.Text != c.text {
			t.Fatalf("%s: text %q want %q", c.hex, v.Text, c.text)
		}
		if c.kind == KindBytes && string(v.Bytes) != c.raw {
			t.Fatalf("%s: bytes mismatch", c.hex)
		}
		enc, err := Encode(v)
		if err != nil || !bytes.Equal(enc, mustHex(c.hex)) {
			t.Fatalf("%s: round-trip %x / %v", c.hex, enc, err)
		}
	}
}

func TestRFCAppendixA_ArraysMaps(t *testing.T) {
	cases := []string{
		"80",                 // []
		"83010203",           // [1,2,3]
		"8301820203820405",   // [1,[2,3],[4,5]]
		"98190102030405060708090a0b0c0d0e0f101112131415161718181819", // [1..25]
		"a0",                 // {}
		"a201020304",         // {1:2,3:4}
		"a26161016162820203", // {"a":1,"b":[2,3]}
		"826161a161626163",   // ["a",{"b":"c"}]
		"a56161614161626142616361436164614461656145", // {"a":"A","b":"B","c":"C","d":"D","e":"E"}
	}
	for _, h := range cases {
		v, err := Decode(mustHex(h))
		if err != nil {
			t.Fatalf("%s: %v", h, err)
		}
		enc, err := Encode(v)
		if err != nil {
			t.Fatalf("%s: encode %v", h, err)
		}
		if !bytes.Equal(enc, mustHex(h)) {
			t.Fatalf("%s: got %x", h, enc)
		}
	}
}

func TestRFCAppendixA_SimpleFloat(t *testing.T) {
	cases := []struct {
		hex string
		check func(t *testing.T, v Value)
	}{
		{"f4", func(t *testing.T, v Value) {
			if v.Kind != KindBool || v.Bool {
				t.Fatal("want false")
			}
		}},
		{"f5", func(t *testing.T, v Value) {
			if v.Kind != KindBool || !v.Bool {
				t.Fatal("want true")
			}
		}},
		{"f6", func(t *testing.T, v Value) {
			if !v.IsNull() {
				t.Fatal("want null")
			}
		}},
		{"f7", func(t *testing.T, v Value) {
			if !v.IsUndefined() {
				t.Fatal("want undefined")
			}
		}},
		{"f90000", func(t *testing.T, v Value) {
			if v.Kind != KindFloat || v.Float != 0 {
				t.Fatalf("got %v", v.Float)
			}
		}},
		{"f93c00", func(t *testing.T, v Value) {
			if v.Float != 1.0 {
				t.Fatalf("got %v", v.Float)
			}
		}},
		{"f93e00", func(t *testing.T, v Value) {
			if v.Float != 1.5 {
				t.Fatalf("got %v", v.Float)
			}
		}},
		{"fa47c35000", func(t *testing.T, v Value) {
			if v.Float != 100000.0 {
				t.Fatalf("got %v", v.Float)
			}
		}},
		{"fb3ff199999999999a", func(t *testing.T, v Value) {
			if math.Abs(v.Float-1.1) > 1e-15 {
				t.Fatalf("got %v", v.Float)
			}
		}},
		{"f97e00", func(t *testing.T, v Value) {
			if !math.IsNaN(v.Float) {
				t.Fatal("want NaN")
			}
		}},
		{"f97c00", func(t *testing.T, v Value) {
			if !math.IsInf(v.Float, 1) {
				t.Fatal("want +Inf")
			}
		}},
		{"f9fc00", func(t *testing.T, v Value) {
			if !math.IsInf(v.Float, -1) {
				t.Fatal("want -Inf")
			}
		}},
	}
	for _, c := range cases {
		v, err := Decode(mustHex(c.hex))
		if err != nil {
			t.Fatalf("%s: %v", c.hex, err)
		}
		c.check(t, v)
	}
}

func TestRFCAppendixA_Tags(t *testing.T) {
	// tag 0 date/time string — stage 1 leaves as Tag
	v, err := Decode(mustHex("c074323031332d30332d32315432303a30343a30305a"))
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != KindTag || v.TagNum != 0 {
		t.Fatalf("got %+v", v)
	}
	if v.TagVal == nil || v.TagVal.Text != "2013-03-21T20:04:00Z" {
		t.Fatalf("tag content: %+v", v.TagVal)
	}

	// tag 1 epoch
	v, err = Decode(mustHex("c11a514b67b0"))
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != KindTag || v.TagNum != 1 {
		t.Fatalf("got %+v", v)
	}
	n, _ := v.TagVal.Int64()
	if n != 1363896240 {
		t.Fatalf("epoch %d", n)
	}
}

func TestDecodeErrors(t *testing.T) {
	bads := []string{
		"",       // empty
		"18",     // truncated uint8 arg
		"19",     // truncated
		"1a0000", // truncated uint32
		"41",     // truncated bytes
		"61",     // truncated text
		"81",     // truncated array
		"a1",     // truncated map
		"c1",     // truncated tag
		"f8",     // truncated simple
		"ff",     // bare break
	}
	for _, h := range bads {
		_, err := Decode(mustHex(h))
		if err == nil {
			t.Fatalf("%s: expected error", h)
		}
	}
}

func TestFromIntEncode(t *testing.T) {
	cases := []int64{0, 1, 23, 24, 255, 256, 65535, 65536, -1, -10, -100, -1000}
	for _, n := range cases {
		v := FromInt(n)
		b, err := Encode(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Decode(b)
		if err != nil {
			t.Fatal(err)
		}
		gn, err := got.Int64()
		if err != nil || gn != n {
			t.Fatalf("%d -> %d (%v)", n, gn, err)
		}
	}
}

func TestNestedDepthLimit(t *testing.T) {
	// deeply nested array [[[[...]]]]
	b := bytes.Repeat([]byte{0x81}, 40)
	b = append(b, 0x00)
	_, err := DecodeWith(b, DecodeOptions{MaxDepth: 10})
	if err == nil {
		t.Fatal("expected depth error")
	}
}

func TestLargeUint(t *testing.T) {
	v, err := Decode(mustHex("1bffffffffffffffff"))
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != KindUnsigned || v.Uint != ^uint64(0) {
		t.Fatalf("%+v", v)
	}
	_, err = v.Int64()
	if err == nil {
		t.Fatal("expected overflow")
	}
}

func TestMapOrderPreserved(t *testing.T) {
	v := MapOf(
		Pair{Key: Text("z"), Value: Uint(1)},
		Pair{Key: Text("a"), Value: Uint(2)},
	)
	b, err := Encode(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Map) != 2 || got.Map[0].Key.Text != "z" || got.Map[1].Key.Text != "a" {
		t.Fatalf("%+v", got.Map)
	}
}

func TestSimpleValues(t *testing.T) {
	for _, s := range []byte{0, 19, 32, 255} {
		v := Simple(s)
		b, err := Encode(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Decode(b)
		if err != nil {
			t.Fatal(err)
		}
		if got.Kind != KindSimple || got.Simple != s {
			t.Fatalf("%d -> %+v", s, got)
		}
	}
}

func TestDecodeAll(t *testing.T) {
	b := append(MustEncode(Uint(1)), MustEncode(Text("hi"))...)
	vs, err := DecodeAll(b, DefaultDecodeOptions())
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 2 || vs[0].Uint != 1 || vs[1].Text != "hi" {
		t.Fatalf("%+v", vs)
	}
}

func TestClone(t *testing.T) {
	v := Array(MapOf(Pair{Key: Text("k"), Value: Bytes([]byte{1, 2})}), Tag(99, Uint(7)))
	c := v.Clone()
	c.Array[0].Map[0].Value.Bytes[0] = 9
	if v.Array[0].Map[0].Value.Bytes[0] != 1 {
		t.Fatal("clone shared bytes")
	}
}
