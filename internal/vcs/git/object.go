package git

import (
	"encoding/hex"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/domain"
)

// 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
// when computing the `git diff` of the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

func decodeOID(sha string) (domain.OID, error) {
	oidBytes, err := hex.DecodeString(sha)
	if err != nil {
		return domain.OID{}, err
	}
	var oid domain.OID
	copy(oid[:], oidBytes)
	return oid, nil
}
