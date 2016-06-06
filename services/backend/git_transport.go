package backend

import (
	"reflect"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	githttp "github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitproto"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/events"
)

// emptyGitCommitID is used in githttp.Event objects in the Last (or
// Commit) field to signify that a branch was created (or deleted).
const emptyGitCommitID = "0000000000000000000000000000000000000000"

func (s *repos) UploadPack(ctx context.Context, op *sourcegraph.UploadPackOp) (*sourcegraph.Packet, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "GitTransport.UploadPack", op.Repo); err != nil {
		// Ignore the error if it is because the repo didn't exist. This comes
		// about when we are implicitly mirroring repos and the metadata is
		// not stored in the database. This is only OK for read access.
		if grpc.Code(err) != codes.NotFound {
			return nil, err
		}
	}

	t, err := store.RepoVCSFromContext(ctx).OpenGitTransport(ctx, op.Repo)
	if err != nil {
		return nil, err
	}

	data, _, err := t.UploadPack(ctx, op.Data, gitproto.TransportOpt{AdvertiseRefs: op.AdvertiseRefs})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.Packet{Data: data}, nil
}

func (s *repos) ReceivePack(ctx context.Context, op *sourcegraph.ReceivePackOp) (*sourcegraph.Packet, error) {
	if err := verifyRepoWriteAccess(ctx, op.Repo); err != nil {
		return nil, err
	}

	t, err := store.RepoVCSFromContext(ctx).OpenGitTransport(ctx, op.Repo)
	if err != nil {
		return nil, err
	}

	data, gitEvents, err := t.ReceivePack(ctx, op.Data, gitproto.TransportOpt{AdvertiseRefs: op.AdvertiseRefs})
	if err != nil {
		return nil, err
	}
	gitEvents = collapseDuplicateEvents(gitEvents)
	payload := events.GitPayload{
		Actor: authpkg.ActorFromContext(ctx).UserSpec(),
		Repo:  op.Repo,
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

	eventsutil.LogGitPush(ctx)

	if err := updateRepoPushedAt(ctx, op.Repo); err != nil {
		return nil, err
	}

	return &sourcegraph.Packet{Data: data}, nil
}

func verifyRepoWriteAccess(ctx context.Context, repoID int32) error {
	repo, err := store.ReposFromContext(ctx).Get(ctx, repoID)
	if err != nil {
		return err
	}
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GitTransport.ReceivePack", repoID); err != nil {
		return err
	}
	if !repo.IsSystemOfRecord() {
		return grpc.Errorf(codes.FailedPrecondition, "repo is not writeable: %s", repo.URI)
	}
	return nil
}

func updateRepoPushedAt(ctx context.Context, repo int32) error {
	now := time.Now()
	return store.ReposFromContext(ctx).Update(ctx, store.RepoUpdate{
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
