package cborwire

import (
	"bytes"
	"sort"
)

// CanonicalEncode encodes v in RFC 8949 Core Deterministic Encoding:
// shortest argument forms, preferred float widths, and map keys sorted
// by their encoded byte order.
func CanonicalEncode(v Value) ([]byte, error) {
	return EncodeWith(v, EncodeOptions{Canonical: true, PreferredFloat: true})
}

func encodeValueCanonical(w writer, v Value, opt EncodeOptions) error {
	switch v.Kind {
	case KindMap:
		type keyed struct {
			keyBytes []byte
			pair     Pair
		}
		items := make([]keyed, len(v.Map))
		for i, p := range v.Map {
			kb, err := EncodeWith(p.Key, EncodeOptions{Canonical: true, PreferredFloat: opt.PreferredFloat})
			if err != nil {
				return err
			}
			items[i] = keyed{keyBytes: kb, pair: p}
		}
		sort.Slice(items, func(i, j int) bool {
			return bytes.Compare(items[i].keyBytes, items[j].keyBytes) < 0
		})
		if err := writeTypeArg(w, MajorMap, uint64(len(items))); err != nil {
			return err
		}
		for _, it := range items {
			if _, err := w.Write(it.keyBytes); err != nil {
				return err
			}
			if err := encodeValueCanonical(w, it.pair.Value, opt); err != nil {
				return err
			}
		}
		return nil
	case KindArray:
		if err := writeTypeArg(w, MajorArray, uint64(len(v.Array))); err != nil {
			return err
		}
		for _, e := range v.Array {
			if err := encodeValueCanonical(w, e, opt); err != nil {
				return err
			}
		}
		return nil
	case KindTag:
		if v.TagVal == nil {
			return encodeValue(w, v, opt)
		}
		if err := writeTypeArg(w, MajorTag, v.TagNum); err != nil {
			return err
		}
		return encodeValueCanonical(w, *v.TagVal, opt)
	default:
		return encodeValue(w, v, opt)
	}
}
