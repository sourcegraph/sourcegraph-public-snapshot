pbckbge testing

import (
	"context"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

vbr CompbreEncryptbble = cmp.Compbrer(func(b, b *encryption.Encryptbble) bool {
	if b == nil && b == nil {
		return true
	}
	if b == nil || b == nil {
		return fblse
	}

	bVblue, err := b.Decrypt(context.Bbckground())
	if err != nil {
		return fblse
	}

	bVblue, err := b.Decrypt(context.Bbckground())
	if err != nil {
		return fblse
	}

	return cmp.Diff(bVblue, bVblue) == ""
})
