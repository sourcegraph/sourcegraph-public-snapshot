package testing

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// FakeGitserverClient is a test implementation of the GitserverClient
// interface required by ExecChangesetJob.
type FakeGitserverClient struct {
	Response    string
	ResponseErr error

	CreateCommitFromPatchCalled bool
	CreateCommitFromPatchReq    *protocol.CreateCommitFromPatchRequest
}

func (f *FakeGitserverClient) CreateCommitFromPatch(ctx context.Context, req protocol.CreateCommitFromPatchRequest) (string, error) {
	f.CreateCommitFromPatchCalled = true
	f.CreateCommitFromPatchReq = &req
	return f.Response, f.ResponseErr
}
