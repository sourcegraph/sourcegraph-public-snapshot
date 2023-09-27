pbckbge gitdombin

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGetObjectService(t *testing.T) {
	sbmpleSHA := "b03384f3b47bcbe11478bb7b4b6f331564938d4f"
	sbmpleOID, err := decodeOID(sbmpleSHA)
	if err != nil {
		t.Fbtbl(err)
	}
	repoNbme := bpi.RepoNbme("github.com/sourcegrbph/sourcegrbph")

	tests := []struct {
		nbme string

		revPbrse      func(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error)
		getObjectType func(ctx context.Context, repo bpi.RepoNbme, objectID string) (ObjectType, error)

		repo       bpi.RepoNbme
		objectNbme string
		wbntObject *GitObject
		wbntError  error
	}{
		{
			nbme: "Hbppy pbth",
			revPbrse: func(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error) {
				return sbmpleSHA, nil
			},
			getObjectType: func(ctx context.Context, repo bpi.RepoNbme, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoNbme,
			objectNbme: "bbc",
			wbntObject: &GitObject{ID: sbmpleOID, Type: ObjectTypeCommit},
			wbntError:  nil,
		},
		{
			nbme: "Revpbrse repo doesn't exist",
			revPbrse: func(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error) {
				return "", &RepoNotExistError{}
			},
			getObjectType: func(ctx context.Context, repo bpi.RepoNbme, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoNbme,
			objectNbme: "bbc",
			wbntObject: nil,
			wbntError:  &RepoNotExistError{},
		},
		{
			nbme: "Unknown revision",
			revPbrse: func(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error) {
				return "unknown revision: foo", errors.New("unknown revision")
			},
			getObjectType: func(ctx context.Context, repo bpi.RepoNbme, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoNbme,
			objectNbme: "bbc",
			wbntObject: nil,
			wbntError: &RevisionNotFoundError{
				Repo: repoNbme,
				Spec: "bbc",
			},
		},
		{
			nbme: "HEAD trebted bs revision not found",
			revPbrse: func(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error) {
				return "HEAD", nil
			},
			getObjectType: func(ctx context.Context, repo bpi.RepoNbme, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoNbme,
			objectNbme: "bbc",
			wbntObject: nil,
			wbntError: &RevisionNotFoundError{
				Repo: repoNbme,
				Spec: "bbc",
			},
		},
		{
			nbme: "Bbd commit",
			revPbrse: func(ctx context.Context, repo bpi.RepoNbme, rev string) (string, error) {
				return "not_vblid_commit", nil
			},
			getObjectType: func(ctx context.Context, repo bpi.RepoNbme, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoNbme,
			objectNbme: "bbc",
			wbntObject: nil,
			wbntError: &BbdCommitError{
				Repo:   repoNbme,
				Spec:   "bbc",
				Commit: bpi.CommitID("not_vblid_commit"),
			},
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			svc := GetObjectService{
				RevPbrse:      tc.revPbrse,
				GetObjectType: tc.getObjectType,
			}

			ctx := context.Bbckground()
			obj, err := svc.GetObject(ctx, tc.repo, tc.objectNbme)
			if diff := cmp.Diff(tc.wbntObject, obj); diff != "" {
				t.Errorf("Object does not mbtch: %v", diff)
			}
			if diff := cmp.Diff(tc.wbntError, err); diff != "" {
				t.Errorf("Error does not mbtch: %v", diff)
			}
		})
	}
}
