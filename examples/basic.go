//go:build ignore

package main

import (
	"fmt"

	"github.com/avohra25/cborwire"
)

func main() {
	v := cborwire.Array(
		cborwire.Uint(1),
		cborwire.Text("cbor"),
		cborwire.MapOf(
			cborwire.Pair{Key: cborwire.Text("ok"), Value: cborwire.Bool(true)},
		),
	)
	b, err := cborwire.Encode(v)
	if err != nil {
		panic(err)
	}
	fmt.Printf("hex: %x\n", b)
	fmt.Printf("diag: %s\n", cborwire.Diagnose(v))

	canon, _ := cborwire.CanonicalEncode(cborwire.MapOf(
		cborwire.Pair{Key: cborwire.Text("z"), Value: cborwire.Uint(1)},
		cborwire.Pair{Key: cborwire.Text("a"), Value: cborwire.Uint(2)},
	))
	fmt.Printf("canonical: %s\n", mustDiag(canon))
}

func mustDiag(b []byte) string {
	s, err := cborwire.DiagnoseBytes(b)
	if err != nil {
		panic(err)
	}
	return s
}
