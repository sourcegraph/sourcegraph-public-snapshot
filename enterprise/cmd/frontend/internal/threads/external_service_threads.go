package threads

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
)

func CreateOnExternalService(ctx context.Context, existingThreadID int64, threadTitle, threadBody, campaignName string, repo *graphqlbackend.RepositoryResolver, patch []byte) (threadID int64, err error) {
	defaultBranch, err := repo.DefaultBranch(ctx)
	if err != nil {
		return 0, err
	}
	oid, err := defaultBranch.Target().OID(ctx)
	if err != nil {
		return 0, err
	}
	var IsAlphanumericWithPeriod = regexp.MustCompile(`[^a-zA-Z0-9_.]+`)
	branchName := "a8n/" + strings.TrimSuffix(IsAlphanumericWithPeriod.ReplaceAllString(campaignName, "-"), "-") // TODO!(sqs): hack

	// TODO!(sqs): For the prototype, prevent changes to any "live" repositories. The sd9 and sd9org
	// namespaces are sandbox/fake accounts used for the prototype.
	if !strings.HasPrefix(repo.Name(), "github.com/sd9/") && !strings.HasPrefix(repo.Name(), "github.com/sd9org/") && !strings.HasPrefix(repo.Name(), "AC/") && !strings.HasPrefix(repo.Name(), "OP/") && !strings.Contains(threadBody, "non-test-repo-ok") {
		return 0, errors.New("refusing to modify non-a8n-test repo")
	}

	// Create a commit and ref.
	refName := "refs/heads/" + branchName
	if _, err := gitserver.DefaultClient.CreateCommitFromPatch(ctx, protocol.CreateCommitFromPatchRequest{
		Repo:       api.RepoName(repo.Name()),
		BaseCommit: api.CommitID(oid),
		TargetRef:  refName,
		Patch:      string(patch),
		CommitInfo: protocol.PatchCommitInfo{
			AuthorName:  "Sourcegraph bot",     // TODO!(sqs): un-hardcode
			AuthorEmail: "bot@sourcegraph.com", // TODO!(sqs): un-hardcode
			Message:     campaignName + " (Sourcegraph campaign)",
			Date:        time.Now(),
		},
	}); err != nil {
		return 0, err
	}

	// Push the newly created ref. TODO!(sqs) this only makes sense for the demo
	cmd := gitserver.DefaultClient.Command("git",
		"-c", "remote.origin.mirror=false", // required to avoid deleting all other local+remote refs not specified in the command
		"push", "-f", "--", "origin",
		refName+":"+refName,
	)
	cmd.Repo = gitserver.Repo{Name: api.RepoName(repo.Name())}
	if out, err := cmd.CombinedOutput(ctx); err != nil {
		return 0, fmt.Errorf("%s\n\n%s", err, out)
	}

	extSvcClient, _, err := getClientForRepo(ctx, repo.DBID())
	if err != nil {
		return 0, errors.WithMessagef(err, "get external service client for repo %d", repo.DBID())
	}

	return extSvcClient.CreateOrUpdateThread(ctx, api.RepoName(repo.Name()), repo.DBID(), repo.DBExternalRepo(), CreateChangesetData{
		BaseRefName:      defaultBranch.AbbrevName(),
		HeadRefName:      branchName,
		Title:            threadTitle,
		Body:             threadBody + fmt.Sprintf("\n\n"+`<img src="https://about.sourcegraph.com/sourcegraph-mark.png" width=12 height=12> Campaign: [%s](#)`, campaignName),
		ExistingThreadID: existingThreadID,
	})
}

func ensureExternalThreadIsPersisted(ctx context.Context, externalThread externalThread, existingThreadID int64) (threadID int64, err error) {
	// If thread exists externally, reuse that.

	// Thread does not yet exist on Sourcegraph.
	thread, err := dbThreads{}.GetByExternal(ctx, externalThread.thread.ExternalServiceID, externalThread.thread.ExternalID)
	if err == nil {
		threadID = thread.ID
	} else if err == errThreadNotFound {
		if existingThreadID == 0 {
			threadID, err = dbCreateExternalThread(ctx, nil, externalThread)
		} else {
			// Thread does exist on Sourcegraph. Link it to the newly created external thread.
			tmp := false
			if _, err := (dbThreads{}).Update(ctx, existingThreadID, dbThreadUpdate{
				BaseRef:    &externalThread.thread.BaseRef,
				BaseRefOID: &externalThread.thread.BaseRefOID,
				HeadRef:    &externalThread.thread.HeadRef,
				HeadRefOID: &externalThread.thread.HeadRefOID,

				IsPendingExternalCreation: &tmp,
				ClearPendingPatch:         true,
				ExternalThreadData:        &externalThread.thread.ExternalThreadData,
			}); err != nil {
				return 0, err
			}
			threadID = existingThreadID
		}
	}
	return threadID, err
}
