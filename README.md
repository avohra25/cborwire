# cborwire

Pure-Go [CBOR](https://www.rfc-editor.org/rfc/rfc8949.html) (RFC 8949 Concise Binary Object Representation) codec toolkit.

**stdlib only** — no third-party modules. Tested with `go test` (and `-race` where practical).

## Features

- Definite- and **indefinite-length** major types (break `0xff`)
- Well-known **tags** 0/1/2/3 (date/time, bignums) with unknown-tag pass-through
- **Canonical / deterministic** encoding (shortest lengths, preferred floats, sorted map keys)
- **Diagnostic notation** (`Diagnose` / `DiagnoseBytes`) per RFC 8949 §8
- **Streaming** `Decoder` over `io.Reader` and `Encoder` to `io.Writer`

Coming next: `cborwire` CLI (`decode` / `encode` / `diagnose` / `validate`) and examples.

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
    v := cborwire.MapOf(
        cborwire.Pair{Key: cborwire.Text("b"), Value: cborwire.Uint(2)},
        cborwire.Pair{Key: cborwire.Text("a"), Value: cborwire.Uint(1)},
    )
    b, _ := cborwire.CanonicalEncode(v)
    fmt.Println(cborwire.DiagnoseBytes(b)) // { "a": 1, "b": 2 } (keys sorted)
}
```

## License

MIT — see [LICENSE](LICENSE).
