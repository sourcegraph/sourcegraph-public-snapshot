package local

import (
	"reflect"

	githttp "github.com/AaronO/go-git-http"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/gitserver/gitpb"
	"src.sourcegraph.com/sourcegraph/pkg/gitproto"
	"src.sourcegraph.com/sourcegraph/server/accesscontrol"
	"src.sourcegraph.com/sourcegraph/store"
)

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
	events.Publish(events.Event{
		EventID: GitPushEvent,
		Payload: GitHookPayload{
			Ctx:    ctx,
			Op:     op,
			Events: gitEvents,
		},
	})
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
