package cborwire

import (
	"math/big"
	"testing"
	"time"
)

func TestTagBignum(t *testing.T) {
	// tag 2: positive bignum 2^64 (bytes 01 00..00 with 8 zero bytes after 01? )
	// RFC: C2 49 010000000000000000 → 18446744073709551616
	v, err := Decode(mustHex("c249010000000000000000"))
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != KindBigInt {
		t.Fatalf("kind %v", v.Kind)
	}
	want := new(big.Int).Lsh(big.NewInt(1), 64)
	if v.BigInt.Cmp(want) != 0 {
		t.Fatalf("got %s want %s", v.BigInt, want)
	}
	enc, err := Encode(v)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	if got.BigInt.Cmp(want) != 0 {
		t.Fatalf("round-trip %s", got.BigInt)
	}

	// tag 3: negative bignum → -18446744073709551617
	v, err = Decode(mustHex("c349010000000000000000"))
	if err != nil {
		t.Fatal(err)
	}
	wantNeg := new(big.Int).Lsh(big.NewInt(1), 64)
	wantNeg.Add(wantNeg, big.NewInt(1)).Neg(wantNeg)
	if v.BigInt.Cmp(wantNeg) != 0 {
		t.Fatalf("got %s want %s", v.BigInt, wantNeg)
	}
}

func TestTagDateTime(t *testing.T) {
	v, err := Decode(mustHex("c074323031332d30332d32315432303a30343a30305a"))
	if err != nil {
		t.Fatal(err)
	}
	tm, err := ParseDateTime(v)
	if err != nil {
		t.Fatal(err)
	}
	if tm.Year() != 2013 || tm.Month() != 3 || tm.Day() != 21 {
		t.Fatalf("%v", tm)
	}

	v, err = Decode(mustHex("c11a514b67b0"))
	if err != nil {
		t.Fatal(err)
	}
	tm, err = ParseDateTime(v)
	if err != nil {
		t.Fatal(err)
	}
	if tm.Unix() != 1363896240 {
		t.Fatalf("%d", tm.Unix())
	}

	now := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	b, err := Encode(DateTimeString(now))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := Decode(b)
	tm, err = ParseDateTime(got)
	if err != nil || tm.Unix() != now.Unix() {
		t.Fatalf("%v vs %v (%v)", tm, now, err)
	}
	b, _ = Encode(EpochDateTime(now))
	got, _ = Decode(b)
	tm, err = ParseDateTime(got)
	if err != nil || tm.Unix() != now.Unix() {
		t.Fatalf("%v %v", tm, err)
	}
}


func TestUnknownTagPassthrough(t *testing.T) {
	v, err := Decode(mustHex("d8184101")) // tag 24, bytes 01 — encoded CBOR tag
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != KindTag || v.TagNum != 24 {
		t.Fatalf("%+v", v)
	}
}

func TestDecodeTagsOff(t *testing.T) {
	v, err := DecodeWith(mustHex("c249010000000000000000"), DecodeOptions{DecodeTags: false, MaxDepth: 32})
	if err != nil {
		t.Fatal(err)
	}
	if v.Kind != KindTag || v.TagNum != 2 {
		t.Fatalf("%+v", v)
	}
}
