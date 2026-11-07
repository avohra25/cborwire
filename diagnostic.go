package cborwire

import (
	"encoding/hex"
	"math/big"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Diagnose returns RFC 8949 §8 diagnostic notation for v.
func Diagnose(v Value) string {
	var b strings.Builder
	writeDiag(&b, v)
	return b.String()
}

func writeDiag(b *strings.Builder, v Value) {
	switch v.Kind {
	case KindUnsigned:
		b.WriteString(strconv.FormatUint(v.Uint, 10))
	case KindNegative:
		// value = -1 - Neg
		n := new(big.Int).SetUint64(v.Neg)
		n.Add(n, big.NewInt(1)).Neg(n)
		b.WriteString(n.String())
	case KindBytes:
		b.WriteString("h'")
		b.WriteString(hex.EncodeToString(v.Bytes))
		b.WriteByte('\'')
	case KindText:
		b.WriteByte('"')
		b.WriteString(escapeDiagText(v.Text))
		b.WriteByte('"')
	case KindArray:
		b.WriteByte('[')
		for i, e := range v.Array {
			if i > 0 {
				b.WriteString(", ")
			}
			writeDiag(b, e)
		}
		b.WriteByte(']')
	case KindMap:
		b.WriteByte('{')
		for i, p := range v.Map {
			if i > 0 {
				b.WriteString(", ")
			}
			writeDiag(b, p.Key)
			b.WriteString(": ")
			writeDiag(b, p.Value)
		}
		b.WriteByte('}')
	case KindTag:
		b.WriteString(strconv.FormatUint(v.TagNum, 10))
		b.WriteByte('(')
		if v.TagVal != nil {
			writeDiag(b, *v.TagVal)
		}
		b.WriteByte(')')
	case KindBool:
		if v.Bool {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case KindNull:
		b.WriteString("null")
	case KindUndefined:
		b.WriteString("undefined")
	case KindFloat:
		writeDiagFloat(b, v.Float)
	case KindSimple:
		b.WriteString("simple(")
		b.WriteString(strconv.Itoa(int(v.Simple)))
		b.WriteByte(')')
	case KindBigInt:
		if v.BigInt == nil {
			b.WriteString("null")
			return
		}
		b.WriteString(v.BigInt.String())
	default:
		b.WriteString(fmt.Sprintf("/*kind:%d*/", v.Kind))
	}
}

func writeDiagFloat(b *strings.Builder, f float64) {
	switch {
	case math.IsNaN(f):
		b.WriteString("NaN")
	case math.IsInf(f, 1):
		b.WriteString("Infinity")
	case math.IsInf(f, -1):
		b.WriteString("-Infinity")
	default:
		b.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
	}
}

func escapeDiagText(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '"', '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				b.WriteString(fmt.Sprintf("\\u%04x", r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// DiagnoseBytes decodes b and returns diagnostic notation.
func DiagnoseBytes(b []byte) (string, error) {
	v, err := Decode(b)
	if err != nil {
		return "", err
	}
	return Diagnose(v), nil
}
