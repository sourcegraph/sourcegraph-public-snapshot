package domain

import (
	"context"
	"encoding/hex"
	"strings"

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
		GetObjectType(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error)
	}
}

func (s *GetObjectService) GetObject(ctx context.Context, repo api.RepoName, objectName string) (*GitObject, error) {
	//if Mocks.GetObject != nil {
	//	return Mocks.GetObject(objectName)
	//}

	// TODO: Maybe we can have a general wrapper around the service
	//span, ctx := ot.StartSpanFromContext(ctx, "Git: GetObject")
	//span.SetTag("objectName", objectName)
	//defer span.Finish()

	if err := checkSpecArgSafety(objectName); err != nil {
		return nil, err
	}

	sha, err := s.RevParser.RevParse(ctx, repo, objectName)
	if err != nil {
		return nil, errors.Wrap(err, "performing rev-parse")
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
