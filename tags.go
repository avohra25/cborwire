package cborwire

import (
	"fmt"
	"math/big"
	"time"
)

// Well-known tag numbers (RFC 8949 §3.4).
const (
	TagDateTimeString uint64 = 0
	TagEpochDateTime  uint64 = 1
	TagPositiveBignum uint64 = 2
	TagNegativeBignum uint64 = 3
)

// expandTag converts well-known tags into richer Values.
func expandTag(num uint64, inner Value) (Value, bool) {
	switch num {
	case TagDateTimeString:
		if inner.Kind != KindText {
			return Value{}, false
		}
		// Keep as Tag but mark decoded; callers can use ParseDateTime.
		return Tag(num, inner), true
	case TagEpochDateTime:
		switch inner.Kind {
		case KindUnsigned, KindNegative, KindFloat:
			return Tag(num, inner), true
		default:
			return Value{}, false
		}
	case TagPositiveBignum:
		if inner.Kind != KindBytes {
			return Value{}, false
		}
		n := new(big.Int).SetBytes(inner.Bytes)
		return BigInt(n), true
	case TagNegativeBignum:
		if inner.Kind != KindBytes {
			return Value{}, false
		}
		n := new(big.Int).SetBytes(inner.Bytes)
		n.Add(n, big.NewInt(1))
		n.Neg(n)
		return BigInt(n), true
	default:
		return Value{}, false
	}
}

// ParseDateTime extracts time.Time from tag 0 (string) or tag 1 (epoch).
func ParseDateTime(v Value) (time.Time, error) {
	if v.Kind != KindTag {
		return time.Time{}, fmt.Errorf("cborwire: not a tag")
	}
	switch v.TagNum {
	case TagDateTimeString:
		if v.TagVal == nil || v.TagVal.Kind != KindText {
			return time.Time{}, fmt.Errorf("cborwire: tag 0 expects text")
		}
		return time.Parse(time.RFC3339Nano, v.TagVal.Text)
	case TagEpochDateTime:
		if v.TagVal == nil {
			return time.Time{}, fmt.Errorf("cborwire: tag 1 empty")
		}
		switch v.TagVal.Kind {
		case KindUnsigned:
			return time.Unix(int64(v.TagVal.Uint), 0).UTC(), nil
		case KindNegative:
			sec := -1 - int64(v.TagVal.Neg)
			return time.Unix(sec, 0).UTC(), nil
		case KindFloat:
			sec := int64(v.TagVal.Float)
			nsec := int64((v.TagVal.Float - float64(sec)) * 1e9)
			return time.Unix(sec, nsec).UTC(), nil
		default:
			return time.Time{}, fmt.Errorf("cborwire: tag 1 unexpected content")
		}
	default:
		return time.Time{}, fmt.Errorf("cborwire: tag %d is not a date/time", v.TagNum)
	}
}

// DateTimeString builds tag 0 from a time.Time.
func DateTimeString(t time.Time) Value {
	return Tag(TagDateTimeString, Text(t.UTC().Format(time.RFC3339Nano)))
}

// EpochDateTime builds tag 1 from a time.Time (integer seconds).
func EpochDateTime(t time.Time) Value {
	return Tag(TagEpochDateTime, FromInt(t.UTC().Unix()))
}
