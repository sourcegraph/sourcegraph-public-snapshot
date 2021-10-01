package domain

import (
	"context"
	"encoding/hex"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

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
		GetObjectType(ctx context.Context, repo api.RepoName, objectID string) (*GitObject, error)
	}
}

func (s *GetObjectService) GetObject(ctx context.Context, repo api.RepoName, objectName string) (GitObject, error) {
	// TODO: Use RevParser and ObjectTyper
	return GitObject{}, errors.New("TODO")
}
