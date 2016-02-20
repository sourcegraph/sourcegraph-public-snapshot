package base

import (
	"crypto/rand"
	"encoding/base32"
	"github.com/golang/glog"
	"io"
	"strings"
)

// Returns garbled strings of printable characters per the base32 spec (though
// we ToLower them for purely aesthetic reasons).
//
// NOTE: lenBytes must be divisible by 5, as otherwise the base32 encoding
// appends information-less "=" characters to pad out the strings it generates.
func SecureRandomBase32(lenBytes int) string {
	if lenBytes%5 != 0 {
		panic("lenBytes must be divisible by 5")
	}

	b := make([]byte, lenBytes)
	_, err := io.ReadFull(rand.Reader, b)
	str := base32.StdEncoding.EncodeToString(b)

	if err != nil {
		glog.Fatalf("SecureRandomBase32 error: %v", err)
	}
	return strings.ToLower(str)
}
