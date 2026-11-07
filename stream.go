package cborwire

import (
	"io"
)

// Decoder reads CBOR data items from an io.Reader.
type Decoder struct {
	r   io.Reader
	opt DecodeOptions
	buf []byte
	off int
	err error
}

// NewDecoder returns a streaming Decoder.
func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderWith(r, DefaultDecodeOptions())
}

// NewDecoderWith returns a Decoder with options.
func NewDecoderWith(r io.Reader, opt DecodeOptions) *Decoder {
	if opt.MaxDepth <= 0 {
		opt.MaxDepth = 32
	}
	return &Decoder{r: r, opt: opt}
}

// Decode reads the next CBOR data item.
func (d *Decoder) Decode() (Value, error) {
	if d.err != nil {
		return Value{}, d.err
	}
	// Fill until we can decode one item; grow buffer as needed.
	for {
		if d.off < len(d.buf) {
			dec := &decoder{buf: d.buf[d.off:], opt: d.opt}
			start := dec.off
			v, err := dec.decodeItem(0)
			if err == nil {
				d.off += dec.off - start
				// compact occasionally
				if d.off > 4096 && d.off > len(d.buf)/2 {
					d.buf = append([]byte(nil), d.buf[d.off:]...)
					d.off = 0
				}
				return v, nil
			}
			if err != io.ErrUnexpectedEOF {
				d.err = err
				return Value{}, err
			}
			// need more data — fall through to read
		}
		tmp := make([]byte, 4096)
		n, err := d.r.Read(tmp)
		if n > 0 {
			d.buf = append(d.buf, tmp[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				if d.off >= len(d.buf) {
					d.err = io.EOF
					return Value{}, io.EOF
				}
				// try once more with what we have
				dec := &decoder{buf: d.buf[d.off:], opt: d.opt}
				v, e2 := dec.decodeItem(0)
				if e2 != nil {
					d.err = e2
					return Value{}, e2
				}
				d.off += dec.off
				return v, nil
			}
			d.err = err
			return Value{}, err
		}
	}
}

// Encoder writes CBOR data items to an io.Writer.
type Encoder struct {
	w   io.Writer
	opt EncodeOptions
}

// NewEncoder returns a streaming Encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// NewEncoderWith returns an Encoder with options.
func NewEncoderWith(w io.Writer, opt EncodeOptions) *Encoder {
	return &Encoder{w: w, opt: opt}
}

// Encode writes one CBOR data item.
func (e *Encoder) Encode(v Value) error {
	var buf []byte
	bw := &byteWriter{buf: &buf}
	var err error
	if e.opt.Canonical {
		err = encodeValueCanonical(bw, v, e.opt)
	} else {
		err = encodeValue(bw, v, e.opt)
	}
	if err != nil {
		return err
	}
	_, err = e.w.Write(buf)
	return err
}
