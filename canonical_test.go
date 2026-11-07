package cborwire

import (
	"bytes"
	"testing"
)

func TestCanonicalMapKeyOrder(t *testing.T) {
	v := MapOf(
		Pair{Key: Text("aa"), Value: Uint(2)},
		Pair{Key: Text("a"), Value: Uint(1)},
		Pair{Key: Uint(1), Value: Uint(3)},
		Pair{Key: Uint(100), Value: Uint(4)},
	)
	b, err := CanonicalEncode(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(b)
	if err != nil {
		t.Fatal(err)
	}
	// Keys sorted by encoded bytes: uint 1 (0x01), uint 100 (0x1864), "a" (0x6161), "aa" (0x626161)
	keys := make([]string, len(got.Map))
	for i, p := range got.Map {
		keys[i] = Diagnose(p.Key)
	}
	wantOrder := []string{"1", "100", `"a"`, `"aa"`}
	for i := range wantOrder {
		if keys[i] != wantOrder[i] {
			t.Fatalf("order %v want %v", keys, wantOrder)
		}
	}
	// re-canonical is stable
	b2, _ := CanonicalEncode(got)
	if !bytes.Equal(b, b2) {
		t.Fatalf("%x vs %x", b, b2)
	}
}

func TestCanonicalShortestInt(t *testing.T) {
	// even without Canonical, writeTypeArg uses shortest; Canonical still works
	for _, n := range []uint64{0, 23, 24, 255, 256, 65535, 65536} {
		b, err := CanonicalEncode(Uint(n))
		if err != nil {
			t.Fatal(err)
		}
		b2, _ := Encode(Uint(n))
		if !bytes.Equal(b, b2) {
			t.Fatalf("%d: %x vs %x", n, b, b2)
		}
	}
}

func TestCanonicalPreferredFloat(t *testing.T) {
	b, err := CanonicalEncode(Float(1.5))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, mustHex("f93e00")) {
		t.Fatalf("got %x want f93e00", b)
	}
}
