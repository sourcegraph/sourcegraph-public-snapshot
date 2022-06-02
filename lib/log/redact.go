package log

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/privacy"
)

func shouldRedact(p privacy.Privacy) bool {
	return p < privacy.Unknown
}

func fnv1a(s string, maxBytes int) uint32 {
	// See https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV_hash_parameters
	hash := uint64(0x811c9dc5)
	for i := 0; i < maxBytes; i++ {
		hash = (hash ^ uint64(s[i])) * 0x01000193
	}
	return uint32(hash)
}

// redact redacts a string and attaches information useful for debugging, including a hash.
//
// The redacted string not have any uniqueness or security guarantees.
func redact(s string) string {
	const maxBytes = 32
	if len(s) > maxBytes {
		return fmt.Sprintf("<redacted:hash=%x,len=%d,hashPrefixLen=32>", fnv1a(s, maxBytes), len(s))
	}
	return fmt.Sprintf("<redacted:hash=%x,len=%d>")
}
