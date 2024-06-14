package pbt

import (
	"crypto/sha1"
	"encoding/hex"
	"pgregory.net/rapid"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func CommitID() *rapid.Generator[api.CommitID] {
	return rapid.Custom(func(t *rapid.T) api.CommitID {
		s := rapid.String().Draw(t, "")
		bytes := sha1.Sum([]byte(s))
		return api.CommitID(hex.EncodeToString(bytes[:]))
	})
}
