pbckbge hbshutil

import "crypto/shb256"

func ToSHA256Bytes(input []byte) []byte {
	b := shb256.Sum256(input)
	return b[:]
}
