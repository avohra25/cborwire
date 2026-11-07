package cborwire

import (
	"bytes"
	"math/rand"
	"testing"
)

func TestRandomRoundTrip(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 200; i++ {
		v := randValue(rng, 3)
		b, err := Encode(v)
		if err != nil {
			t.Fatalf("%d encode: %v", i, err)
		}
		got, err := Decode(b)
		if err != nil {
			t.Fatalf("%d decode: %v (%x)", i, err, b)
		}
		b2, err := Encode(got)
		if err != nil {
			t.Fatalf("%d re-encode: %v", i, err)
		}
		if !bytes.Equal(b, b2) {
			t.Fatalf("%d unstable: %x vs %x diag=%s", i, b, b2, Diagnose(v))
		}
	}
}

func TestCanonicalIdempotent(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 50; i++ {
		v := randValue(rng, 2)
		b1, err := CanonicalEncode(v)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Decode(b1)
		if err != nil {
			t.Fatal(err)
		}
		b2, err := CanonicalEncode(got)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b1, b2) {
			t.Fatalf("canonical not idempotent %x vs %x", b1, b2)
		}
	}
}

func randValue(rng *rand.Rand, depth int) Value {
	if depth <= 0 {
		switch rng.Intn(6) {
		case 0:
			return Uint(rng.Uint64() % 1000)
		case 1:
			return FromInt(rng.Int63n(200) - 100)
		case 2:
			return Text(randString(rng, 8))
		case 3:
			b := make([]byte, rng.Intn(8))
			rng.Read(b)
			return Bytes(b)
		case 4:
			return Bool(rng.Intn(2) == 0)
		default:
			return Null()
		}
	}
	switch rng.Intn(8) {
	case 0:
		return Uint(uint64(rng.Intn(1 << 20)))
	case 1:
		return FromInt(int64(rng.Intn(10000) - 5000))
	case 2:
		return Text(randString(rng, 12))
	case 3:
		b := make([]byte, rng.Intn(16))
		rng.Read(b)
		return Bytes(b)
	case 4:
		n := rng.Intn(4)
		arr := make([]Value, n)
		for i := range arr {
			arr[i] = randValue(rng, depth-1)
		}
		return Array(arr...)
	case 5:
		n := rng.Intn(4)
		pairs := make([]Pair, n)
		for i := range pairs {
			pairs[i] = Pair{Key: Text(randString(rng, 4)), Value: randValue(rng, depth-1)}
		}
		return MapOf(pairs...)
	case 6:
		return Tag(uint64(rng.Intn(40)), randValue(rng, depth-1))
	default:
		return Float(float64(rng.Intn(100)) / 2)
	}
}

func randString(rng *rand.Rand, n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, rng.Intn(n)+1)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}
