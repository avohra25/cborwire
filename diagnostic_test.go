package cborwire

import (
	"math"
	"testing"
)

func TestDiagnoseRFC(t *testing.T) {
	cases := []struct {
		hex, want string
	}{
		{"00", "0"},
		{"01", "1"},
		{"0a", "10"},
		{"17", "23"},
		{"1818", "24"},
		{"20", "-1"},
		{"29", "-10"},
		{"f4", "false"},
		{"f5", "true"},
		{"f6", "null"},
		{"f7", "undefined"},
		{"40", "h''"},
		{"4401020304", "h'01020304'"},
		{"60", `""`},
		{"6161", `"a"`},
		{"6449455446", `"IETF"`},
		{"80", "[]"},
		{"83010203", "[1, 2, 3]"},
		{"a0", "{}"},
		{"a201020304", "{1: 2, 3: 4}"},
		{"c074323031332d30332d32315432303a30343a30305a", `0("2013-03-21T20:04:00Z")`},
		{"f93e00", "1.5"},
		{"f97c00", "Infinity"},
		{"f9fc00", "-Infinity"},
		{"f97e00", "NaN"},
		{"f98000", "-0"},
	}
	for _, c := range cases {
		got, err := DiagnoseBytes(mustHex(c.hex))
		if err != nil {
			t.Fatalf("%s: %v", c.hex, err)
		}
		if c.want == "NaN" {
			if got != "NaN" {
				t.Fatalf("%s: got %q", c.hex, got)
			}
			continue
		}
		if got != c.want {
			t.Fatalf("%s: got %q want %q", c.hex, got, c.want)
		}
	}
	_ = math.NaN
}

func TestDiagnoseNested(t *testing.T) {
	v := Array(MapOf(Pair{Key: Text("k"), Value: Bool(true)}), Null())
	s := Diagnose(v)
	if s != `[{"k": true}, null]` {
		t.Fatalf("%q", s)
	}
}
