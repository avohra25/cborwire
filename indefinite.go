package cborwire

// EncodeIndefinite encodes arrays/maps/bytes/text using indefinite-length
// form (break stop code 0xff). Other kinds fall back to definite Encode.
func EncodeIndefinite(v Value) ([]byte, error) {
	var buf []byte
	w := &byteWriter{buf: &buf}
	if err := encodeIndef(w, v, EncodeOptions{}); err != nil {
		return nil, err
	}
	return buf, nil
}

func encodeIndef(w writer, v Value, opt EncodeOptions) error {
	switch v.Kind {
	case KindBytes:
		if err := w.WriteByte(0x5f); err != nil {
			return err
		}
		if len(v.Bytes) > 0 {
			// single definite chunk
			if err := encodeValue(w, Bytes(v.Bytes), opt); err != nil {
				return err
			}
		}
		return w.WriteByte(0xff)
	case KindText:
		if err := w.WriteByte(0x7f); err != nil {
			return err
		}
		if v.Text != "" {
			if err := encodeValue(w, Text(v.Text), opt); err != nil {
				return err
			}
		}
		return w.WriteByte(0xff)
	case KindArray:
		if err := w.WriteByte(0x9f); err != nil {
			return err
		}
		for _, e := range v.Array {
			if err := encodeValue(w, e, opt); err != nil {
				return err
			}
		}
		return w.WriteByte(0xff)
	case KindMap:
		if err := w.WriteByte(0xbf); err != nil {
			return err
		}
		for _, p := range v.Map {
			if err := encodeValue(w, p.Key, opt); err != nil {
				return err
			}
			if err := encodeValue(w, p.Value, opt); err != nil {
				return err
			}
		}
		return w.WriteByte(0xff)
	default:
		return encodeValue(w, v, opt)
	}
}
