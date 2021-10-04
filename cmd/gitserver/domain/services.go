package domain

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

// TODO: This is our domain package, we don't want it importing anything
// TODO: This whole package should probably live in internal/gitserver/domain so that cmd/gitserver becomes thinner

// OID is a Git OID (40-char hex-encoded).
type OID [20]byte

func (oid OID) String() string { return hex.EncodeToString(oid[:]) }

// ObjectType is a valid Git object type (commit, tag, tree, and blob).
type ObjectType string

// Standard Git object types.
const (
	ObjectTypeCommit ObjectType = "commit"
	ObjectTypeTag    ObjectType = "tag"
	ObjectTypeTree   ObjectType = "tree"
	ObjectTypeBlob   ObjectType = "blob"
)

// GitObject represents a GitObject
type GitObject struct {
	ID   OID
	Type ObjectType
}

type GetObjectService struct {
	RevParser interface {
		RevParse(ctx context.Context, repo api.RepoName, rev string) (string, error)
	}
	ObjectTyper interface {
		GetObjectType(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error)
	}
}

func (s *GetObjectService) GetObject(ctx context.Context, repo api.RepoName, objectName string) (*GitObject, error) {
	// TODO: We shouldn't need this mock since we can instead mock out the adapters
	//if Mocks.GetObject != nil {
	//	return Mocks.GetObject(objectName)
	//}

	// TODO: Maybe we can have a general wrapper around the service. Tracing
	// shouldn't be something the domain package is concerned by.

	//span, ctx := ot.StartSpanFromContext(ctx, "Git: GetObject")
	//span.SetTag("objectName", objectName)
	//defer span.Finish()

	if err := checkSpecArgSafety(objectName); err != nil {
		return nil, err
	}

	sha, err := s.RevParser.RevParse(ctx, repo, objectName)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return nil, err
		}
		if strings.Contains(sha, "unknown revision") {
			return nil, &RevisionNotFoundError{Repo: repo, Spec: objectName}
		}
		return nil, err
	}

	commit := strings.TrimSpace(sha)
	if !IsAbsoluteRevision(commit) {
		if commit == "HEAD" {
			// We don't verify the existence of HEAD, but if HEAD doesn't point to anything
			// git just returns `HEAD` as the output of rev-parse. An example where this
			// occurs is an empty repository.
			return nil, &RevisionNotFoundError{Repo: repo, Spec: objectName}
		}
		return nil, BadCommitError{Spec: objectName, Commit: api.CommitID(commit), Repo: repo}
	}

	oid, err := decodeOID(sha)
	if err != nil {
		return nil, errors.Wrap(err, "decoding oid")
	}

	objectType, err := s.ObjectTyper.GetObjectType(ctx, repo, oid.String())
	if err != nil {
		return nil, errors.Wrap(err, "getting object type")
	}

	return &GitObject{
		ID:   oid,
		Type: objectType,
	}, nil
}

// checkSpecArgSafety returns a non-nil err if spec begins with a "-", which could
// cause it to be interpreted as a git command line argument.
func checkSpecArgSafety(spec string) error {
	if strings.HasPrefix(spec, "-") {
		return errors.Errorf("invalid git revision spec %q (begins with '-')", spec)
	}
	return nil
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

// IsAbsoluteRevision checks if the revision is a git OID SHA string.
//
// Note: This doesn't mean the SHA exists in a repository, nor does it mean it
// isn't a ref. Git allows 40-char hexadecimal strings to be references.
func IsAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(('0' <= r && r <= '9') ||
			('a' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return false
		}
	}
	return true
}
