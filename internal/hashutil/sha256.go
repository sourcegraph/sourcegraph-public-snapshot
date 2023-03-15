package hashutil

import "crypto/sha256"

func ToSHA256Bytes(input []byte) []byte {
	b := sha256.Sum256(input)
	return b[:]
}
