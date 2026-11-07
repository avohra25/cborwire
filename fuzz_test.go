package cborwire

import (
	"bytes"
	"testing"
)

func FuzzDecodeEncode(f *testing.F) {
	seeds := []string{
		"00", "20", "40", "60", "80", "a0", "f4", "f5", "f6", "f7",
		"83010203", "a201020304", "f93e00", "fb3ff199999999999a",
		"5f42010243030405ff", "9f010203ff", "bf61610161629f0203ffff",
		"c074323031332d30332d32315432303a30343a30305a",
		"c249010000000000000000",
		"1818", "1903e8", "1a000f4240",
		"6449455446", "4401020304",
	}
	for _, s := range seeds {
		f.Add(mustHex(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			return
		}
		v, err := DecodeWith(data, DecodeOptions{MaxDepth: 16, MaxBytes: 4096, DecodeTags: true})
		if err != nil {
			return
		}
		b, err := Encode(v)
		if err != nil {
			t.Fatal(err)
		}
		v2, err := Decode(b)
		if err != nil {
			t.Fatal(err)
		}
		b2, err := Encode(v2)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, b2) {
			t.Fatalf("unstable encode")
		}
		_ = Diagnose(v2)
		_, _ = CanonicalEncode(v2)
	})
}
