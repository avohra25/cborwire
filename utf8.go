package cborwire

import "unicode/utf8"

func decodeRuneImpl(p []byte) (rune, int) {
	return utf8.DecodeRune(p)
}
