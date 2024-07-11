package testing

import (
	"context"
	"encoding/hex"
	"io"
	"math/rand"
	"strings"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

type MockedChangesetSyncState struct {
	// DiffStat is the diff.Stat of the mocked "git diff" call to gitserver.
	DiffStat *diff.Stat

	MockClient *gitserver.MockClient
}

// MockChangesetSyncState sets up mocks such that invoking SetDerivedState() with
// a Changeset will use the same diff (+1, ~1, -3) when setting the SyncState
// on a Changeset.
//
// state.Unmock() must called to clean up, usually via defer.
func MockChangesetSyncState(repo *protocol.RepoInfo) *MockedChangesetSyncState {
	state := &MockedChangesetSyncState{
		// This diff.Stat matches the testGitHubDiff below
		DiffStat: &diff.Stat{Added: 2, Deleted: 4},
	}

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.DiffFunc.SetDefaultHook(func(_ context.Context, repo api.RepoName, opts gitserver.DiffOptions) (*gitserver.DiffFileIterator, error) {
		// This provides a diff that will resolve to 1 added line, 1 changed
		// line, and 3 deleted lines.
		const testGitHubDiff = `
diff --git a/test.py b/test.py
index 884601b..c4886d5 100644
--- a/test.py
+++ b/test.py
@@ -1,6 +1,4 @@
+# square makes a value squarer.
 def square(a):
-    """
-    square makes a value squarer.
-    """

-    return a * a
+    return pow(a, 2)

`
		return gitserver.NewDiffFileIterator(io.NopCloser(strings.NewReader(testGitHubDiff))), nil
	})
	gitserverClient.ResolveRevisionFunc.SetDefaultHook(func(context.Context, api.RepoName, string, gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID(generateFakeCommitID()), nil
	})

	state.MockClient = gitserverClient
	return state
}

func generateFakeCommitID() string {
	// Generate a random byte slice with 20 bytes (160 bits)
	commitBytes := make([]byte, 20)
	_, err := rand.Read(commitBytes)
	if err != nil {
		panic(err)
	}

	// Convert the byte slice to a hexadecimal string
	commitID := hex.EncodeToString(commitBytes)

	return commitID
}
