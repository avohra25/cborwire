# cborwire

Pure-Go [CBOR](https://www.rfc-editor.org/rfc/rfc8949.html) (RFC 8949 Concise Binary Object Representation) codec toolkit.

**stdlib only** — no third-party modules. Tested with `go test` (and `-race` where practical).

## Features

- Definite-length major types: unsigned/negative ints, bytes, text, arrays, maps, tags, simples, half/float32/float64
- **Indefinite-length** byte/text strings, arrays and maps (break `0xff`)
- Well-known **tags**: 0 (date/time string), 1 (epoch), 2/3 (bignums → `math/big`), unknown tags pass through
- Helpers: `DateTimeString`, `EpochDateTime`, `ParseDateTime`, `EncodeIndefinite`

Coming next: canonical encoding, diagnostic notation, streaming API, CLI.

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
