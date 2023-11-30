package git

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGetObject(t *testing.T) {
	sampleSHA := "a03384f3a47acae11478ba7b4a6f331564938d4f"
	sampleOID, err := decodeOID(sampleSHA)
	if err != nil {
		t.Fatal(err)
	}
	repoName := api.RepoName("github.com/sourcegraph/sourcegraph")

	tests := []struct {
		name string

		revParse      revParseFunc
		getObjectType getObjectTypeFunc

		repo       api.RepoName
		objectName string
		wantObject *gitdomain.GitObject
		wantError  error
	}{
		{
			name: "Happy path",
			revParse: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, rev string) (string, error) {
				return sampleSHA, nil
			},
			getObjectType: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, objectID string) (gitdomain.ObjectType, error) {
				return gitdomain.ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: &gitdomain.GitObject{ID: sampleOID, Type: gitdomain.ObjectTypeCommit},
			wantError:  nil,
		},
		{
			name: "Revparse repo doesn't exist",
			revParse: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, rev string) (string, error) {
				return "", &gitdomain.RepoNotExistError{}
			},
			getObjectType: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, objectID string) (gitdomain.ObjectType, error) {
				return gitdomain.ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError:  &gitdomain.RepoNotExistError{},
		},
		{
			name: "Unknown revision",
			revParse: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, rev string) (string, error) {
				return "unknown revision: foo", errors.New("unknown revision")
			},
			getObjectType: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, objectID string) (gitdomain.ObjectType, error) {
				return gitdomain.ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError: &gitdomain.RevisionNotFoundError{
				Repo: repoName,
				Spec: "abc",
			},
		},
		{
			name: "HEAD treated as revision not found",
			revParse: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, rev string) (string, error) {
				return "HEAD", nil
			},
			getObjectType: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, objectID string) (gitdomain.ObjectType, error) {
				return gitdomain.ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError: &gitdomain.RevisionNotFoundError{
				Repo: repoName,
				Spec: "abc",
			},
		},
		{
			name: "Bad commit",
			revParse: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, rev string) (string, error) {
				return "not_valid_commit", nil
			},
			getObjectType: func(ctx context.Context, rcf *wrexec.RecordingCommandFactory, repo api.RepoName, dir common.GitDir, objectID string) (gitdomain.ObjectType, error) {
				return gitdomain.ObjectTypeCommit, nil
			},
			repo:       repoName,
			objectName: "abc",
			wantObject: nil,
			wantError: &gitdomain.BadCommitError{
				Repo:   repoName,
				Spec:   "abc",
				Commit: api.CommitID("not_valid_commit"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			obj, err := getObject(ctx, wrexec.NewNoOpRecordingCommandFactory(), t.TempDir(), tc.getObjectType, tc.revParse, tc.repo, tc.objectName)
			if diff := cmp.Diff(tc.wantObject, obj); diff != "" {
				t.Errorf("Object does not match: %v", diff)
			}
			if diff := cmp.Diff(tc.wantError, err); diff != "" {
				t.Errorf("Error does not match: %v", diff)
			}
		})
	}
}
