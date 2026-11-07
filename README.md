# cborwire

Pure-Go [CBOR](https://www.rfc-editor.org/rfc/rfc8949.html) (RFC 8949 Concise Binary Object Representation) codec toolkit.

**stdlib only** — no third-party modules. Tested with `go test` (and `-race` where practical).

## Status

Stage 1 — definite-length major types:

- Unsigned / negative integers (all additional-info widths)
- Byte strings and UTF-8 text strings
- Arrays and maps (order-preserving)
- Tags (pass-through; well-known tags expanded in later stages)
- Simple values: `false` / `true` / `null` / `undefined` / unassigned simples
- Floats: half (float16), float32, float64

Coming next: indefinite-length items, tag 0/1/2/3 helpers, canonical encoding, diagnostic notation, streaming `io.Reader`/`io.Writer` API, and a `cborwire` CLI.

## Install

```bash
go get github.com/avohra25/cborwire@latest
```

## Quick start

```go
package main

import (
    "fmt"
    "github.com/avohra25/cborwire"
)

func main() {
    v := cborwire.Array(
        cborwire.Uint(1),
        cborwire.Text("hi"),
        cborwire.Bool(true),
    )
    b, err := cborwire.Encode(v)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%x\n", b) // 8301626869f5

    out, err := cborwire.Decode(b)
    if err != nil {
        panic(err)
    }
    fmt.Println(out.Array[1].Text) // hi
}
```

## License

MIT — see [LICENSE](LICENSE).
