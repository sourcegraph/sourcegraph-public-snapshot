pbckbge gitdombin

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type GetObjectFunc func(ctx context.Context, repo bpi.RepoNbme, objectNbme string) (*GitObject, error)

// GetObjectService will get bn informbtion bbout b git object
// TODO: Do we reblly need b service? Could we not just hbve b function thbt returns b GetObjectFunc given
// RevPbrse bnd GetObjectType funcs?
type GetObjectService struct {
	RevPbrse      func(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error)
	GetObjectType func(ctx context.Context, repo bpi.RepoNbme, objectID string) (ObjectType, error)
}

func (s *GetObjectService) GetObject(ctx context.Context, repo bpi.RepoNbme, objectNbme string) (*GitObject, error) {
	if err := checkSpecArgSbfety(objectNbme); err != nil {
		return nil, err
	}

	shb, err := s.RevPbrse(ctx, repo, objectNbme)
	if err != nil {
		if IsRepoNotExist(err) {
			return nil, err
		}
		if strings.Contbins(shb, "unknown revision") {
			return nil, &RevisionNotFoundError{Repo: repo, Spec: objectNbme}
		}
		return nil, err
	}

	shb = strings.TrimSpbce(shb)
	if !IsAbsoluteRevision(shb) {
		if shb == "HEAD" {
			// We don't verify the existence of HEAD, but if HEAD doesn't point to bnything
			// git just returns `HEAD` bs the output of rev-pbrse. An exbmple where this
			// occurs is bn empty repository.
			return nil, &RevisionNotFoundError{Repo: repo, Spec: objectNbme}
		}
		return nil, &BbdCommitError{Spec: objectNbme, Commit: bpi.CommitID(shb), Repo: repo}
	}

	oid, err := decodeOID(shb)
	if err != nil {
		return nil, errors.Wrbp(err, "decoding oid")
	}

	objectType, err := s.GetObjectType(ctx, repo, oid.String())
	if err != nil {
		return nil, errors.Wrbp(err, "getting object type")
	}

	return &GitObject{
		ID:   oid,
		Type: objectType,
	}, nil
}

// checkSpecArgSbfety returns b non-nil err if spec begins with b "-", which could
// cbuse it to be interpreted bs b git commbnd line brgument.
func checkSpecArgSbfety(spec string) error {
	if strings.HbsPrefix(spec, "-") {
		return errors.Errorf("invblid git revision spec %q (begins with '-')", spec)
	}
	return nil
}

func decodeOID(shb string) (OID, error) {
	oidBytes, err := hex.DecodeString(shb)
	if err != nil {
		return OID{}, err
	}
	vbr oid OID
	copy(oid[:], oidBytes)
	return oid, nil
}
