package local

import (
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	githttp "github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/eventsutil"
)

// emptyGitCommitID is used in githttp.Event objects in the Last (or
// Commit) field to signify that a branch was created (or deleted).
const emptyGitCommitID = "0000000000000000000000000000000000000000"

var GitTransport gitpb.GitTransportServer = &gitTransport{}

type gitTransport struct{}

var _ gitpb.GitTransportServer = (*gitTransport)(nil)

func (s *gitTransport) InfoRefs(ctx context.Context, op *gitpb.InfoRefsOp) (*gitpb.Packet, error) {
	// This service is read-only, but can be followed by a write
	// action. If we only deny access once writing it leads to a confusing user
	// experience
	if op.Service == gitproto.ReceivePack {
		if err := verifyRepoWriteAccess(ctx, op.Repo); err != nil {
			return nil, err
		}
	}

	store := store.RepoVCSFromContext(ctx)
	t, err := store.OpenGitTransport(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	data, err := t.InfoRefs(ctx, op.Service)
	if err != nil {
		return nil, err
	}
	return &gitpb.Packet{Data: data}, nil
}

func (s *gitTransport) UploadPack(ctx context.Context, op *gitpb.UploadPackOp) (*gitpb.Packet, error) {
	store := store.RepoVCSFromContext(ctx)
	t, err := store.OpenGitTransport(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	data, _, err := t.UploadPack(ctx, op.Data, gitproto.TransportOpt{ContentEncoding: op.ContentEncoding})
	if err != nil {
		return nil, err
	}
	return &gitpb.Packet{Data: data}, nil
}

func (s *gitTransport) ReceivePack(ctx context.Context, op *gitpb.ReceivePackOp) (*gitpb.Packet, error) {
	if err := verifyRepoWriteAccess(ctx, op.Repo); err != nil {
		return nil, err
	}

	t, err := store.RepoVCSFromContext(ctx).OpenGitTransport(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	data, gitEvents, err := t.ReceivePack(ctx, op.Data, gitproto.TransportOpt{ContentEncoding: op.ContentEncoding})
	if err != nil {
		return nil, err
	}
	gitEvents = collapseDuplicateEvents(gitEvents)
	payload := events.GitPayload{
		Actor:           authpkg.UserSpecFromContext(ctx),
		Repo:            op.Repo,
		ContentEncoding: op.ContentEncoding,
	}
	for _, e := range gitEvents {
		payload.Event = e
		if e.Last == emptyGitCommitID {
			events.Publish(events.GitCreateBranchEvent, payload)
		} else if e.Commit == emptyGitCommitID {
			events.Publish(events.GitDeleteBranchEvent, payload)
		} else if e.Type == githttp.PUSH || e.Type == githttp.PUSH_FORCE || e.Type == githttp.TAG {
			events.Publish(events.GitPushEvent, payload)
		}
	}

	eventsutil.LogHTTPGitPush(ctx)

	if err := updateRepoPushedAt(ctx, op.Repo); err != nil {
		return nil, err
	}

	return &gitpb.Packet{Data: data}, nil
}

func verifyRepoWriteAccess(ctx context.Context, repoSpec sourcegraph.RepoSpec) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GitTransport.ReceivePack"); err != nil {
		return err
	}
	repo, err := svc.Repos(ctx).Get(ctx, &repoSpec)
	if err != nil {
		return err
	}
	if !repo.IsSystemOfRecord() {
		return grpc.Errorf(codes.FailedPrecondition, "repo is not writeable %v", repoSpec.URI)
	}
	return nil
}

func updateRepoPushedAt(ctx context.Context, repo sourcegraph.RepoSpec) error {
	now := time.Now()
	return store.ReposFromContext(ctx).Update(ctx, &store.RepoUpdate{
		ReposUpdateOp: &sourcegraph.ReposUpdateOp{Repo: repo},
		PushedAt:      &now,
		// Note: No need to update the UpdatedAt field, since it
		// should track significant updates to repo metadata, not just
		// pushes.
	})
}

// collapseDuplicateEvents transforms a githttp event list such that adjacent
// equivalent events are collapsed into a single event.
func collapseDuplicateEvents(eventsDup []githttp.Event) []githttp.Event {
	events := []githttp.Event{}
	var previousEvent githttp.Event
	for _, e := range eventsDup {
		if !reflect.DeepEqual(e, previousEvent) {
			events = append(events, e)
		}
		previousEvent = e
	}
	return events
}
