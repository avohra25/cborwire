# cborwire

Pure-Go [CBOR](https://www.rfc-editor.org/rfc/rfc8949.html) (RFC 8949 Concise Binary Object Representation) codec toolkit.

**stdlib only** — no third-party modules. Tested with `go test`, `go test -race`, and `go test -fuzz`.

## Features

| Area | Support |
|------|---------|
| Major types | Unsigned / negative ints, byte & text strings, arrays, maps, tags, simples, half / float32 / float64 |
| Indefinite length | Arrays, maps, bytes, text with break `0xff` (`EncodeIndefinite` + decode) |
| Tags | 0 (date/time string), 1 (epoch), 2/3 (bignums → `math/big`); unknown tags pass through |
| Canonical | Shortest lengths, preferred floats, map keys sorted by encoded bytes (`CanonicalEncode`) |
| Diagnostic | RFC 8949 §8 notation via `Diagnose` / `DiagnoseBytes` |
| Streaming | `NewDecoder(io.Reader)` / `NewEncoder(io.Writer)` |
| CLI | `decode` · `encode` · `diagnose` · `validate` |

## Install

```bash
go get github.com/avohra25/cborwire@latest
go install github.com/avohra25/cborwire/cmd/cborwire@latest
```

## Library quick start

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
    fmt.Printf("%x\n", b)                 // 8301626869f5
    fmt.Println(cborwire.Diagnose(v))     // [1, "hi", true]

    canon, _ := cborwire.CanonicalEncode(cborwire.MapOf(
        cborwire.Pair{Key: cborwire.Text("b"), Value: cborwire.Uint(2)},
        cborwire.Pair{Key: cborwire.Text("a"), Value: cborwire.Uint(1)},
    ))
    fmt.Println(cborwire.MustDiagnose(canon)) // {"a": 1, "b": 2}
}
```

## CLI

```bash
# Hex → diagnostic notation
cborwire diagnose 8301626869f5
# → [1, "hi", true]

# Hex → JSON-ish
cborwire decode -out json 8301626869f5

# JSON → CBOR hex (canonical)
cborwire encode -canonical '{"a":1,"b":2}'

# Validate
cborwire validate -format hex 9f010203ff
```

Input may be a CLI argument, `-file path`, or stdin. Use `-format hex|base64|raw`.

## Examples

```bash
go run examples/basic.go
```

## Tests

```bash
go test ./...
go test -race ./...
go test -fuzz=FuzzDecodeEncode -fuzztime=10s
```

## License

MIT — see [LICENSE](LICENSE).
