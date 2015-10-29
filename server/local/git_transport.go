package local

import (
	"reflect"

	githttp "github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	"src.sourcegraph.com/sourcegraph/notif/githooks"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

// githttp.Event objects have the EmptyCommitID value in the Last (or Commit)
// field to signify that a branch was created (or deleted).
const EmptyCommitID = "0000000000000000000000000000000000000000"

var GitTransport gitpb.GitTransportServer = &gitTransport{}

type gitTransport struct{}

var _ gitpb.GitTransportServer = (*gitTransport)(nil)

func (s *gitTransport) InfoRefs(ctx context.Context, op *gitpb.InfoRefsOp) (*gitpb.Packet, error) {
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
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GitTransport.ReceivePack"); err != nil {
		return nil, err
	}

	store := store.RepoVCSFromContext(ctx)
	t, err := store.OpenGitTransport(ctx, op.Repo.URI)
	if err != nil {
		return nil, err
	}

	data, gitEvents, err := t.ReceivePack(ctx, op.Data, gitproto.TransportOpt{ContentEncoding: op.ContentEncoding})
	if err != nil {
		return nil, err
	}
	gitEvents = collapseDuplicateEvents(gitEvents)
	payload := githooks.Payload{
		CtxActor:        authpkg.ActorFromContext(ctx),
		Repo:            op.Repo,
		ContentEncoding: op.ContentEncoding,
	}
	for _, e := range gitEvents {
		payload.Event = e
		if e.Last == EmptyCommitID {
			payload.Type = githooks.GitCreateEvent
			events.Publish(events.Event{
				EventID: githooks.GitCreateEvent,
				Payload: payload,
			})
		} else if e.Commit == EmptyCommitID {
			payload.Type = githooks.GitDeleteEvent
			events.Publish(events.Event{
				EventID: githooks.GitDeleteEvent,
				Payload: payload,
			})
		} else if e.Type == githttp.PUSH || e.Type == githttp.PUSH_FORCE {
			payload.Type = githooks.GitPushEvent
			events.Publish(events.Event{
				EventID: githooks.GitPushEvent,
				Payload: payload,
			})
		}
	}
	return &gitpb.Packet{Data: data}, nil
}

// collapseDuplicateEvents transforms a githttp event list such that adjacent
// equivalent events are collapsed into a single event
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
