package testing

import (
	"context"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

var CompareEncryptable = cmp.Comparer(func(a, b *encryption.Encryptable) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	aValue, err := a.Decrypted(context.Background())
	if err != nil {
		return false
	}

	bValue, err := b.Decrypted(context.Background())
	if err != nil {
		return false
	}

	return cmp.Diff(aValue, bValue) == ""
})
