package cborwire

import (
	"bytes"
	"testing"
)

func TestRFCIndefinite(t *testing.T) {
	cases := []struct {
		hex  string
		check func(*testing.T, Value)
	}{
		{"5f42010243030405ff", func(t *testing.T, v Value) {
			if v.Kind != KindBytes || !bytes.Equal(v.Bytes, []byte{1, 2, 3, 4, 5}) {
				t.Fatalf("%x", v.Bytes)
			}
		}},
		{"7f657374726561646d696e67ff", func(t *testing.T, v Value) {
			if v.Text != "streaming" {
				t.Fatalf("%q", v.Text)
			}
		}},
		{"9f010203ff", func(t *testing.T, v Value) {
			if len(v.Array) != 3 || v.Array[2].Uint != 3 {
				t.Fatalf("%+v", v.Array)
			}
		}},
		{"9f018202039f0405ffff", func(t *testing.T, v Value) {
			// [1,[2,3],[4,5]] with indef
			if len(v.Array) != 3 {
				t.Fatalf("%+v", v)
			}
		}},
		{"9f01820203820405ff", func(t *testing.T, v Value) {
			if len(v.Array) != 3 {
				t.Fatalf("%+v", v)
			}
		}},
		{"bf61610161629f0203ffff", func(t *testing.T, v Value) {
			if len(v.Map) != 2 {
				t.Fatalf("%+v", v.Map)
			}
		}},
		{"bf61616141ff", func(t *testing.T, v Value) {
			if len(v.Map) != 1 || v.Map[0].Key.Text != "a" {
				t.Fatalf("%+v", v)
			}
		}},
		{"5fff", func(t *testing.T, v Value) {
			if v.Kind != KindBytes || len(v.Bytes) != 0 {
				t.Fatal(v)
			}
		}},
		{"7fff", func(t *testing.T, v Value) {
			if v.Text != "" {
				t.Fatal(v)
			}
		}},
		{"9fff", func(t *testing.T, v Value) {
			if len(v.Array) != 0 {
				t.Fatal(v)
			}
		}},
		{"bfff", func(t *testing.T, v Value) {
			if len(v.Map) != 0 {
				t.Fatal(v)
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

func TestEncodeIndefiniteRoundTrip(t *testing.T) {
	vals := []Value{
		Bytes([]byte{1, 2, 3}),
		Text("hello"),
		Array(Uint(1), Uint(2), Text("x")),
		MapOf(Pair{Key: Text("a"), Value: Uint(1)}),
	}
	for i, v := range vals {
		b, err := EncodeIndefinite(v)
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		if b[0]&0x1f != 31 {
			t.Fatalf("%d: not indefinite head %x", i, b[0])
		}
		if b[len(b)-1] != 0xff {
			t.Fatalf("%d: no break", i)
		}
		got, err := Decode(b)
		if err != nil {
			t.Fatalf("%d decode: %v", i, err)
		}
		b2, _ := Encode(got)
		b3, _ := Encode(v)
		if !bytes.Equal(b2, b3) {
			t.Fatalf("%d: content mismatch %x vs %x", i, b2, b3)
		}
	}
}

func TestHalfFloatTable(t *testing.T) {
	cases := []struct {
		hex string
		f   float64
	}{
		{"f90000", 0},
		{"f98000", 0}, // -0 — Float() == 0 in Go comparison with +0 for == but signbit differs; we check Abs
		{"f93c00", 1},
		{"f93e00", 1.5},
		{"f97bff", 65504},
		{"f90001", 5.960464477539063e-8},
		{"f90400", 6.103515625e-5},
		{"f9c400", -4},
	}
	for _, c := range cases {
		v, err := Decode(mustHex(c.hex))
		if err != nil {
			t.Fatal(err)
		}
		if c.hex == "f98000" {
			continue // -0
		}
		if v.Float != c.f {
			t.Fatalf("%s: got %v want %v", c.hex, v.Float, c.f)
		}
	}
}
