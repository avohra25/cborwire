package cborwire

import (
	"bytes"
	"io"
	"testing"
)

func TestStreamRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	vals := []Value{Uint(42), Text("hi"), Array(Bool(true), Null()), Float(1.5)}
	for _, v := range vals {
		if err := enc.Encode(v); err != nil {
			t.Fatal(err)
		}
	}
	dec := NewDecoder(&buf)
	for i, want := range vals {
		got, err := dec.Decode()
		if err != nil {
			t.Fatalf("%d: %v", i, err)
		}
		b1, _ := Encode(want)
		b2, _ := Encode(got)
		if !bytes.Equal(b1, b2) {
			t.Fatalf("%d: %x vs %x", i, b1, b2)
		}
	}
	_, err := dec.Decode()
	if err != io.EOF {
		t.Fatalf("want EOF got %v", err)
	}
}

func TestStreamCanonicalEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoderWith(&buf, EncodeOptions{Canonical: true, PreferredFloat: true})
	v := MapOf(
		Pair{Key: Text("b"), Value: Uint(2)},
		Pair{Key: Text("a"), Value: Uint(1)},
	)
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
	ref, _ := CanonicalEncode(v)
	if !bytes.Equal(buf.Bytes(), ref) {
		t.Fatalf("%x vs %x", buf.Bytes(), ref)
	}
}

func TestStreamPartialReads(t *testing.T) {
	payload := MustEncode(Array(Uint(1), Text("abcdefghij"), Bytes(bytes.Repeat([]byte{9}, 100))))
	r := &slowReader{data: payload, n: 3}
	dec := NewDecoder(r)
	got, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}
	if got.Kind != KindArray || len(got.Array) != 3 {
		t.Fatalf("%+v", got)
	}
}

type slowReader struct {
	data []byte
	n    int
	off  int
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.off >= len(s.data) {
		return 0, io.EOF
	}
	n := s.n
	if n > len(p) {
		n = len(p)
	}
	if s.off+n > len(s.data) {
		n = len(s.data) - s.off
	}
	copy(p, s.data[s.off:s.off+n])
	s.off += n
	return n, nil
}
