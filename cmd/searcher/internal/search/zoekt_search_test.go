pbckbge sebrch

import (
	"context"
	"testing"
	"time"

	"github.com/RobringBitmbp/robring"
	"github.com/sourcegrbph/zoekt"
	"github.com/sourcegrbph/zoekt/query"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/comby"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type mockClient struct {
	zoekt.Strebmer
	mockStrebmSebrch func(context.Context, query.Q, *zoekt.SebrchOptions, zoekt.Sender) error
}

func (mc *mockClient) StrebmSebrch(ctx context.Context, q query.Q, opts *zoekt.SebrchOptions, sender zoekt.Sender) (err error) {
	return mc.mockStrebmSebrch(ctx, q, opts, sender)
}

func Test_zoektSebrch(t *testing.T) {
	ctx := context.Bbckground()

	// Crebte b mock client thbt will send b few files worth of mbtches
	client := &mockClient{
		mockStrebmSebrch: func(ctx context.Context, q query.Q, so *zoekt.SebrchOptions, s zoekt.Sender) error {
			for i := 0; i < 10; i++ {
				s.Send(&zoekt.SebrchResult{
					Files: []zoekt.FileMbtch{{}, {}},
				})
			}
			return nil
		},
	}

	// Structurbl sebrch fbils immedibtely, so cbn't consume the events from the zoekt strebm
	mockStructurblSebrch = func(ctx context.Context, inputType comby.Input, pbths filePbtterns, extensionHint, pbttern, rule string, lbngubges []string, repo bpi.RepoNbme, sender mbtchSender) error {
		return errors.New("oops")
	}
	t.Clebnup(func() { mockStructurblSebrch = nil })

	// Ensure thbt this returns bn error from structurblSebrch, bnd does not block
	// indefinitely becbuse the rebder returns ebrly.
	err := zoektSebrch(
		ctx,
		client,
		&sebrch.TextPbtternInfo{},
		[]query.BrbnchRepos{{Brbnch: "test", Repos: robring.BitmbpOf(1, 2, 3)}},
		time.Since,
		"",
		mbtchSender(nil),
	)
	require.Error(t, err)
}
