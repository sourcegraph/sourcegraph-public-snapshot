package git

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// OID is a Git OID (40-char hex-encoded).
type OID [20]byte

func (oid OID) String() string { return hex.EncodeToString(oid[:]) }

// 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
// when computing the `git diff` of the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// ObjectType is a valid Git object type (commit, tag, tree, and blob).
type ObjectType string

// Standard Git object types.
const (
	ObjectTypeCommit ObjectType = "commit"
	ObjectTypeTag    ObjectType = "tag"
	ObjectTypeTree   ObjectType = "tree"
	ObjectTypeBlob   ObjectType = "blob"
)

// GetObject looks up a Git object and returns information about it.
func GetObject(ctx context.Context, repo api.RepoName, objectName string) (oid OID, objectType ObjectType, err error) {
	if Mocks.GetObject != nil {
		return Mocks.GetObject(objectName)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: GetObject")
	span.SetTag("objectName", objectName)
	defer span.Finish()

	if err := checkSpecArgSafety(objectName); err != nil {
		return oid, "", err
	}

	cmd := gitserver.DefaultClient.Command("git", "rev-parse", objectName)
	cmd.Repo = repo
	sha, err := runRevParse(ctx, cmd, objectName)
	if err != nil {
		return oid, "", err
	}
	oid, err = decodeOID(string(sha))
	if err != nil {
		return oid, "", err
	}

	// Check the SHA is safe (as an extra precaution).
	if err := checkSpecArgSafety(string(sha)); err != nil {
		return oid, "", err
	}
	cmd = gitserver.DefaultClient.Command("git", "cat-file", "-t", "--", string(sha))
	cmd.Repo = repo
	out, err := cmd.Output(ctx)
	if err != nil {
		return oid, "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", cmd.Args, out))
	}
	objectType = ObjectType(string(bytes.TrimSpace(out)))
	return oid, objectType, nil
}

func decodeOID(sha string) (OID, error) {
	oidBytes, err := hex.DecodeString(sha)
	if err != nil {
		return OID{}, err
	}
	var oid OID
	copy(oid[:], oidBytes)
	return oid, nil
}
