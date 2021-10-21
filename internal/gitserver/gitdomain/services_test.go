package gitdomain

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestGetObjectService(t *testing.T) {
	sampleSHA := "a03384f3a47acae11478ba7b4a6f331564938d4f"
	sampleOID, err := decodeOID(sampleSHA)
	if err != nil {
		t.Fatal(err)
	}
	repoName := api.RepoName("github.com/sourcegraph/sourcegraph")

	tests := []struct {
		name string

		revParse      func(ctx context.Context, repo api.RepoName, rev string) (string, error)
		getObjectType func(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error)

		repo       api.RepoName
		objectName string
		wantObject *GitObject
		wantError  error
	}{
		{
			name: "Happy path",
			revParse: func(ctx context.Context, repo api.RepoName, rev string) (string, error) {
				return sampleSHA, nil
			},
			getObjectType: func(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: &GitObject{ID: sampleOID, Type: ObjectTypeCommit},
			wantError:  nil,
		},
		{
			name: "Revparse repo doesn't exist",
			revParse: func(ctx context.Context, repo api.RepoName, rev string) (string, error) {
				return "", &RepoNotExistError{}
			},
			getObjectType: func(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError:  &RepoNotExistError{},
		},
		{
			name: "Unknown revision",
			revParse: func(ctx context.Context, repo api.RepoName, rev string) (string, error) {
				return "unknown revision: foo", errors.New("unknown revision")
			},
			getObjectType: func(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError: &RevisionNotFoundError{
				Repo: repoName,
				Spec: "abc",
			},
		},
		{
			name: "HEAD treated as revision not found",
			revParse: func(ctx context.Context, repo api.RepoName, rev string) (string, error) {
				return "HEAD", nil
			},
			getObjectType: func(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError: &RevisionNotFoundError{
				Repo: repoName,
				Spec: "abc",
			},
		},
		{
			name: "Bad commit",
			revParse: func(ctx context.Context, repo api.RepoName, rev string) (string, error) {
				return "not_valid_commit", nil
			},
			getObjectType: func(ctx context.Context, repo api.RepoName, objectID string) (ObjectType, error) {
				return ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError: BadCommitError{
				Repo:   repoName,
				Spec:   "abc",
				Commit: api.CommitID("not_valid_commit"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := GetObjectService{
				RevParse:      tc.revParse,
				GetObjectType: tc.getObjectType,
			}

			ctx := context.Background()
			obj, err := svc.GetObject(ctx, tc.repo, tc.objectName)
			if diff := cmp.Diff(tc.wantObject, obj); diff != "" {
				t.Errorf("Object does not match: %v", diff)
			}
			if diff := cmp.Diff(tc.wantError, err); diff != "" {
				t.Errorf("Error does not match: %v", diff)
			}
		})
	}

}
