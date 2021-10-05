package git

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/domain"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t tree /dev/null`, which is used as the base
// when computing the `git diff` of the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// GetObject looks up a Git object and returns information about it.
// TODO: This should be replaced by gitserver.Client.GetObject
func GetObject(ctx context.Context, repo api.RepoName, objectName string) (oid domain.OID, objectType domain.ObjectType, err error) {
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
	objectType = domain.ObjectType(bytes.TrimSpace(out))
	return oid, objectType, nil
}

func decodeOID(sha string) (domain.OID, error) {
	oidBytes, err := hex.DecodeString(sha)
	if err != nil {
		return domain.OID{}, err
	}
	var oid domain.OID
	copy(oid[:], oidBytes)
	return oid, nil
}
