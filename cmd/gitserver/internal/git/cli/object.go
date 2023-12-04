package cli

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (g *gitCLIBackend) GetObject(ctx context.Context, objectName string) (*gitdomain.GitObject, error) {
	return getObject(ctx, g.rcf, g.dir, getObjectType, revParse, g.repoName, objectName)
}

func getObject(ctx context.Context, rcf *wrexec.RecordingCommandFactory, dir common.GitDir, getObjectType getObjectTypeFunc, revParse revParseFunc, repo api.RepoName, objectName string) (_ *gitdomain.GitObject, err error) {
	tr, ctx := trace.New(ctx, "GetObject",
		attribute.String("objectName", objectName))
	defer tr.EndWithErr(&err)

	if err := git.CheckSpecArgSafety(objectName); err != nil {
		return nil, err
	}

	sha, err := revParse(ctx, rcf, repo, dir, objectName)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return nil, err
		}
		if strings.Contains(sha, "unknown revision") {
			return nil, &gitdomain.RevisionNotFoundError{Repo: repo, Spec: objectName}
		}
		return nil, err
	}

	sha = strings.TrimSpace(sha)
	if !gitdomain.IsAbsoluteRevision(sha) {
		if sha == "HEAD" {
			// We don't verify the existence of HEAD, but if HEAD doesn't point to anything
			// git just returns `HEAD` as the output of rev-parse. An example where this
			// occurs is an empty repository.
			return nil, &gitdomain.RevisionNotFoundError{Repo: repo, Spec: objectName}
		}
		return nil, &gitdomain.BadCommitError{Spec: objectName, Commit: api.CommitID(sha), Repo: repo}
	}

	oid, err := decodeOID(sha)
	if err != nil {
		return nil, errors.Wrap(err, "decoding oid")
	}

	objectType, err := getObjectType(ctx, rcf, repo, dir, oid.String())
	if err != nil {
		return nil, errors.Wrap(err, "getting object type")
	}

	return &gitdomain.GitObject{
		ID:   oid,
		Type: objectType,
	}, nil

}

type getObjectTypeFunc func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, objectID string) (gitdomain.ObjectType, error)

// getObjectType returns the object type given an objectID.
func getObjectType(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, objectID string) (gitdomain.ObjectType, error) {
	cmd := exec.Command("git", "cat-file", "-t", "--", objectID)
	dir.Set(cmd)
	wrappedCmd := rcf.WrapWithRepoName(context.Background(), log.NoOp(), repo, cmd)
	out, err := wrappedCmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", wrappedCmd.Args, out))
	}

	objectType := gitdomain.ObjectType(bytes.TrimSpace(out))
	return objectType, nil
}

type revParseFunc func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, rev string) (string, error)

// revParse will run rev-parse on the given rev.
func revParse(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, rev string) (string, error) {
	cmd := exec.Command("git", "rev-parse", rev)
	dir.Set(cmd)
	wrappedCmd := rcf.WrapWithRepoName(context.Background(), log.NoOp(), repo, cmd)
	out, err := wrappedCmd.CombinedOutput()
	if err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (output: %q)", wrappedCmd.Args, out))
	}

	return string(out), nil
}

func decodeOID(sha string) (gitdomain.OID, error) {
	oidBytes, err := hex.DecodeString(sha)
	if err != nil {
		return gitdomain.OID{}, err
	}
	var oid gitdomain.OID
	copy(oid[:], oidBytes)
	return oid, nil
}
