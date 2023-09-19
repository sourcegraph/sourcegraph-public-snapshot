package tst

import (
	"encoding/base64"
	"math"

	"github.com/google/uuid"
)

func id() string {
	id := []byte(uuid.NewString())
	return base64.RawStdEncoding.EncodeToString(id[:])

}

func joinID(v, sep, id string, max int) string {
	length := int(math.Min(float64(len(id)), float64(max-len(sep)-len(v))))
	return v + sep + id[:length]
}

func boolp(v bool) *bool {
	return &v
}

func strp(v string) *string {
	return &v
}

func reverse[T any](src []T) []T {
	reversed := make([]T, 0, len(src))
	for i := len(src) - 1; i >= 0; i-- {
		reversed = append(reversed, src[i])
	}
	return reversed
}
