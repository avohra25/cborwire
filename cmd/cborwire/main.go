// Command cborwire is a CLI for the cborwire CBOR toolkit.
package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/avohra25/cborwire"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "decode":
		cmdDecode(os.Args[2:])
	case "encode":
		cmdEncode(os.Args[2:])
	case "diagnose":
		cmdDiagnose(os.Args[2:])
	case "validate":
		cmdValidate(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `cborwire — pure-Go CBOR (RFC 8949) toolkit

Usage:
  cborwire decode   [flags] [input]
  cborwire encode   [flags] [input]
  cborwire diagnose [flags] [input]
  cborwire validate [flags] [input]

decode flags:
  -format hex|base64|raw   input encoding (default hex)
  -out    diag|json|hex    output form (default diag)
  -file   path             read input from file (else arg or stdin)

encode flags:
  -canonical               use deterministic encoding
  -out    hex|base64|raw   output encoding (default hex)

diagnose / validate: same -format/-file as decode.
`)
}

func readInput(file, arg string) ([]byte, error) {
	if file != "" {
		return os.ReadFile(file)
	}
	if arg != "" {
		return []byte(arg), nil
	}
	return io.ReadAll(os.Stdin)
}

func parseCBORBytes(raw []byte, format string) ([]byte, error) {
	s := strings.TrimSpace(string(raw))
	switch format {
	case "hex":
		s = strings.ReplaceAll(s, " ", "")
		s = strings.ReplaceAll(s, "\n", "")
		return hex.DecodeString(s)
	case "base64":
		return base64.StdEncoding.DecodeString(s)
	case "raw":
		return raw, nil
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}
}

func cmdDecode(args []string) {
	fs := flag.NewFlagSet("decode", flag.ExitOnError)
	format := fs.String("format", "hex", "")
	out := fs.String("out", "diag", "")
	file := fs.String("file", "", "")
	_ = fs.Parse(args)
	arg := ""
	if fs.NArg() > 0 {
		arg = fs.Arg(0)
	}
	raw, err := readInput(*file, arg)
	check(err)
	b, err := parseCBORBytes(raw, *format)
	check(err)
	v, err := cborwire.Decode(b)
	check(err)
	switch *out {
	case "diag":
		fmt.Println(cborwire.Diagnose(v))
	case "hex":
		enc, err := cborwire.Encode(v)
		check(err)
		fmt.Println(hex.EncodeToString(enc))
	case "json":
		j, err := valueToJSON(v)
		check(err)
		enc, err := json.MarshalIndent(j, "", "  ")
		check(err)
		fmt.Println(string(enc))
	default:
		check(fmt.Errorf("unknown -out %q", *out))
	}
}

func cmdDiagnose(args []string) {
	fs := flag.NewFlagSet("diagnose", flag.ExitOnError)
	format := fs.String("format", "hex", "")
	file := fs.String("file", "", "")
	_ = fs.Parse(args)
	arg := ""
	if fs.NArg() > 0 {
		arg = fs.Arg(0)
	}
	raw, err := readInput(*file, arg)
	check(err)
	b, err := parseCBORBytes(raw, *format)
	check(err)
	s, err := cborwire.DiagnoseBytes(b)
	check(err)
	fmt.Println(s)
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	format := fs.String("format", "hex", "")
	file := fs.String("file", "", "")
	_ = fs.Parse(args)
	arg := ""
	if fs.NArg() > 0 {
		arg = fs.Arg(0)
	}
	raw, err := readInput(*file, arg)
	check(err)
	b, err := parseCBORBytes(raw, *format)
	check(err)
	_, err = cborwire.Decode(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("ok")
}

func cmdEncode(args []string) {
	fs := flag.NewFlagSet("encode", flag.ExitOnError)
	canonical := fs.Bool("canonical", false, "")
	out := fs.String("out", "hex", "")
	file := fs.String("file", "", "")
	_ = fs.Parse(args)
	arg := ""
	if fs.NArg() > 0 {
		arg = fs.Arg(0)
	}
	raw, err := readInput(*file, arg)
	check(err)
	var j any
	check(json.Unmarshal(raw, &j))
	v, err := jsonToValue(j)
	check(err)
	var b []byte
	if *canonical {
		b, err = cborwire.CanonicalEncode(v)
	} else {
		b, err = cborwire.Encode(v)
	}
	check(err)
	switch *out {
	case "hex":
		fmt.Println(hex.EncodeToString(b))
	case "base64":
		fmt.Println(base64.StdEncoding.EncodeToString(b))
	case "raw":
		_, _ = os.Stdout.Write(b)
	default:
		check(fmt.Errorf("unknown -out %q", *out))
	}
}

func jsonToValue(j any) (cborwire.Value, error) {
	switch x := j.(type) {
	case nil:
		return cborwire.Null(), nil
	case bool:
		return cborwire.Bool(x), nil
	case float64:
		if x == float64(int64(x)) && x >= -1e15 && x <= 1e15 {
			return cborwire.FromInt(int64(x)), nil
		}
		return cborwire.Float(x), nil
	case string:
		return cborwire.Text(x), nil
	case []any:
		arr := make([]cborwire.Value, len(x))
		for i, e := range x {
			v, err := jsonToValue(e)
			if err != nil {
				return cborwire.Value{}, err
			}
			arr[i] = v
		}
		return cborwire.Array(arr...), nil
	case map[string]any:
		pairs := make([]cborwire.Pair, 0, len(x))
		for k, val := range x {
			vv, err := jsonToValue(val)
			if err != nil {
				return cborwire.Value{}, err
			}
			pairs = append(pairs, cborwire.Pair{Key: cborwire.Text(k), Value: vv})
		}
		return cborwire.MapOf(pairs...), nil
	default:
		return cborwire.Value{}, fmt.Errorf("unsupported JSON type %T", j)
	}
}

func valueToJSON(v cborwire.Value) (any, error) {
	switch v.Kind {
	case cborwire.KindUnsigned:
		return v.Uint, nil
	case cborwire.KindNegative:
		n, err := v.Int64()
		if err != nil {
			bi, err2 := v.AsBigInt()
			if err2 != nil {
				return nil, err
			}
			return bi.String(), nil
		}
		return n, nil
	case cborwire.KindBytes:
		return map[string]any{"$bytes": hex.EncodeToString(v.Bytes)}, nil
	case cborwire.KindText:
		return v.Text, nil
	case cborwire.KindArray:
		out := make([]any, len(v.Array))
		for i, e := range v.Array {
			j, err := valueToJSON(e)
			if err != nil {
				return nil, err
			}
			out[i] = j
		}
		return out, nil
	case cborwire.KindMap:
		out := make(map[string]any, len(v.Map))
		for _, p := range v.Map {
			var key string
			switch p.Key.Kind {
			case cborwire.KindText:
				key = p.Key.Text
			default:
				key = cborwire.Diagnose(p.Key)
			}
			jv, err := valueToJSON(p.Value)
			if err != nil {
				return nil, err
			}
			out[key] = jv
		}
		return out, nil
	case cborwire.KindBool:
		return v.Bool, nil
	case cborwire.KindNull:
		return nil, nil
	case cborwire.KindUndefined:
		return map[string]any{"$undefined": true}, nil
	case cborwire.KindFloat:
		return v.Float, nil
	case cborwire.KindTag:
		inner, err := valueToJSON(*v.TagVal)
		if err != nil {
			return nil, err
		}
		return map[string]any{"$tag": v.TagNum, "$value": inner}, nil
	case cborwire.KindSimple:
		return map[string]any{"$simple": v.Simple}, nil
	case cborwire.KindBigInt:
		return v.BigInt.String(), nil
	default:
		return nil, fmt.Errorf("unsupported kind %d", v.Kind)
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
