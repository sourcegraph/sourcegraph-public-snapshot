pbckbge stbte

import (
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func TestCblcCounts(t *testing.T) {
	t.Pbrbllel()

	now := timeutil.Now()
	dbysAgo := func(dbys int) time.Time { return now.AddDbte(0, 0, -dbys) }

	tests := []struct {
		codehosts  string
		nbme       string
		chbngesets []*btypes.Chbngeset
		stbrt      time.Time
		end        time.Time
		events     []*btypes.ChbngesetEvent
		wbnt       []*ChbngesetCounts
	}{
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open merged",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			nbme: "stbrt end time on subset of events",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			end:   dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset crebted bnd closed before stbrt time",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(8)),
			},
			stbrt: dbysAgo(4),
			end:   dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(7), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 1, Merged: 1},
				{Time: dbysAgo(3), Totbl: 1, Merged: 1},
				{Time: dbysAgo(2), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset crebted bnd closed before stbrt time",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(8)),
			},
			stbrt: dbysAgo(4),
			end:   dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(7), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 1, Merged: 1},
				{Time: dbysAgo(3), Totbl: 1, Merged: 1},
				{Time: dbysAgo(2), Totbl: 1, Merged: 1},
			},
		},
		{
			nbme: "stbrt time not even x*24hours before end time",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(3),
			end:   now.Add(-18 * time.Hour),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 0, Merged: 0},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "multiple chbngesets open merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
				ghChbngeset(2, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 2),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 2, Open: 2, OpenPending: 2},
				{Time: dbysAgo(1), Totbl: 2, Merged: 2},
				{Time: dbysAgo(0), Totbl: 2, Merged: 2},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "multiple chbngesets open merged",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(2)),
				bbsChbngeset(2, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 2),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 2, Open: 2, OpenPending: 2},
				{Time: dbysAgo(1), Totbl: 2, Merged: 2},
				{Time: dbysAgo(0), Totbl: 2, Merged: 2},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "multiple chbngesets open merged different times",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
				ghChbngeset(2, dbysAgo(2)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubMerged, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 2),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: dbysAgo(1), Totbl: 2, Merged: 2},
				{Time: dbysAgo(0), Totbl: 2, Merged: 2},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "multiple chbngesets open merged different times",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
				bbsChbngeset(2, dbysAgo(2)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(2), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 2),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: dbysAgo(1), Totbl: 2, Merged: 2},
				{Time: dbysAgo(0), Totbl: 2, Merged: 2},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "chbngeset merged bnd closed bt sbme time",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubClosed, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "chbngeset merged bnd closed bt sbme time, reversed order in slice",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open closed reopened merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(4)),
			},
			stbrt: dbysAgo(5),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(3), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubReopened, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(5), Totbl: 0, Open: 0},
				{Time: dbysAgo(4), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(3), Totbl: 1, Open: 0, Closed: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open declined reopened merged",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(4)),
			},
			stbrt: dbysAgo(5),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(3), btypes.ChbngesetEventKindBitbucketServerDeclined, 1),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindBitbucketServerReopened, 1),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(5), Totbl: 0, Open: 0},
				{Time: dbysAgo(4), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(3), Totbl: 1, Open: 0, Closed: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "multiple chbngesets open closed reopened merged different times",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(5)),
				ghChbngeset(2, dbysAgo(4)),
			},
			stbrt: dbysAgo(6),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(4), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindGitHubClosed, 2),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindGitHubReopened, 1),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubReopened, 2),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindGitHubMerged, 2),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(6), Totbl: 0, Open: 0},
				{Time: dbysAgo(5), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(4), Totbl: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: dbysAgo(3), Totbl: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: dbysAgo(2), Totbl: 2, Open: 2, OpenPending: 2},
				{Time: dbysAgo(1), Totbl: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 2, Merged: 2},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "multiple chbngesets open declined reopened merged different times",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(5)),
				bbsChbngeset(2, dbysAgo(4)),
			},
			stbrt: dbysAgo(6),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(4), btypes.ChbngesetEventKindBitbucketServerDeclined, 1),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindBitbucketServerDeclined, 2),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindBitbucketServerReopened, 1),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindBitbucketServerReopened, 2),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindBitbucketServerMerged, 2),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(6), Totbl: 0, Open: 0},
				{Time: dbysAgo(5), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(4), Totbl: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: dbysAgo(3), Totbl: 2, Open: 1, OpenPending: 1, Closed: 1},
				{Time: dbysAgo(2), Totbl: 2, Open: 2, OpenPending: 2},
				{Time: dbysAgo(1), Totbl: 2, Open: 1, OpenPending: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 2, Merged: 2},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open closed reopened merged, unsorted events",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(4)),
			},
			stbrt: dbysAgo(5),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubReopened, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(5), Totbl: 0, Open: 0},
				{Time: dbysAgo(4), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(3), Totbl: 1, Open: 0, Closed: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open closed reopened merged, unsorted events",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(4)),
			},
			stbrt: dbysAgo(5),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindBitbucketServerDeclined, 1),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindBitbucketServerReopened, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(5), Totbl: 0, Open: 0},
				{Time: dbysAgo(4), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(3), Totbl: 1, Open: 0, Closed: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open, bpproved, merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "APPROVED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, bpproved, merged",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, chbnges-requested, unbpproved",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsPbrticipbntEvent(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerDismissed),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open, bpproved, closed, reopened",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "APPROVED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindGitHubReopened, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, bpproved, declined, reopened",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerDeclined, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindBitbucketServerReopened, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open, bpproved, closed, merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "APPROVED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindGitHubReopened, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, bpproved, closed, merged",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerDeclined, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindBitbucketServerReopened, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open, chbnges-requested, closed, reopened",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "CHANGES_REQUESTED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindGitHubReopened, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, chbnges-requested, closed, reopened",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerDeclined, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindBitbucketServerReopened, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open, chbnges-requested, closed, merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "CHANGES_REQUESTED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open, comment review, bpproved, merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(3), "user1", "COMMENTED"),
				ghReview(1, dbysAgo(2), "user2", "APPROVED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, comment review, bpproved, merged",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(3), "user1", btypes.ChbngesetEventKindBitbucketServerCommented),
				bbsActivity(1, dbysAgo(2), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset multiple bpprovbls counting once",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(1), "user1", "APPROVED"),
				ghReview(1, dbysAgo(0), "user2", "APPROVED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset multiple bpprovbls counting once",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(0), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset multiple chbnges-requested reviews counting once",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(1, dbysAgo(0), "user2", "CHANGES_REQUESTED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 0, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 0, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset multiple chbnges-requested reviews counting once",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(1, dbysAgo(0), "user2", btypes.ChbngesetEventKindBitbucketServerReviewed),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 0, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 0, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset open, chbnges-requested, merged",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "CHANGES_REQUESTED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, chbnges-requested, merged",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Merged: 1},
				{Time: dbysAgo(0), Totbl: 1, Merged: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "multiple chbngesets open different review stbges before merge",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(6)),
				ghChbngeset(2, dbysAgo(6)),
				ghChbngeset(3, dbysAgo(6)),
			},
			stbrt: dbysAgo(7),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(5), "user1", "APPROVED"),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindGitHubMerged, 1),
				ghReview(2, dbysAgo(4), "user1", "APPROVED"),
				ghReview(2, dbysAgo(3), "user2", "APPROVED"),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubMerged, 2),
				ghReview(3, dbysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(3, dbysAgo(1), "user2", "CHANGES_REQUESTED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 3),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(7), Totbl: 0, Open: 0},
				{Time: dbysAgo(6), Totbl: 3, Open: 3, OpenPending: 3},
				{Time: dbysAgo(5), Totbl: 3, Open: 3, OpenPending: 2, OpenApproved: 1},
				{Time: dbysAgo(4), Totbl: 3, Open: 3, OpenPending: 1, OpenApproved: 2},
				{Time: dbysAgo(3), Totbl: 3, Open: 2, OpenPending: 1, OpenApproved: 1, Merged: 1},
				{Time: dbysAgo(2), Totbl: 3, Open: 1, OpenPending: 0, OpenChbngesRequested: 1, Merged: 2},
				{Time: dbysAgo(1), Totbl: 3, Merged: 3},
				{Time: dbysAgo(0), Totbl: 3, Merged: 3},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "multiple chbngesets open different review stbges before merge",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(6)),
				bbsChbngeset(2, dbysAgo(6)),
				bbsChbngeset(3, dbysAgo(6)),
			},
			stbrt: dbysAgo(7),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(5), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindBitbucketServerMerged, 1),
				bbsActivity(2, dbysAgo(4), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(2, dbysAgo(3), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindBitbucketServerMerged, 2),
				bbsActivity(3, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(3, dbysAgo(1), "user2", btypes.ChbngesetEventKindBitbucketServerReviewed),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 3),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(7), Totbl: 0, Open: 0},
				{Time: dbysAgo(6), Totbl: 3, Open: 3, OpenPending: 3},
				{Time: dbysAgo(5), Totbl: 3, Open: 3, OpenPending: 2, OpenApproved: 1},
				{Time: dbysAgo(4), Totbl: 3, Open: 3, OpenPending: 1, OpenApproved: 2},
				{Time: dbysAgo(3), Totbl: 3, Open: 2, OpenPending: 1, OpenApproved: 1, Merged: 1},
				{Time: dbysAgo(2), Totbl: 3, Open: 1, OpenPending: 0, OpenChbngesRequested: 1, Merged: 2},
				{Time: dbysAgo(1), Totbl: 3, Merged: 3},
				{Time: dbysAgo(0), Totbl: 3, Merged: 3},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "time slice of multiple chbngesets in different stbges before merge",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(6)),
				ghChbngeset(2, dbysAgo(6)),
				ghChbngeset(3, dbysAgo(6)),
			},
			// Sbme test bs bbove, except we only look bt 3 dbys in the middle
			stbrt: dbysAgo(4),
			end:   dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(5), "user1", "APPROVED"),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindGitHubMerged, 1),
				ghReview(2, dbysAgo(4), "user1", "APPROVED"),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubMerged, 2),
				ghReview(3, dbysAgo(2), "user1", "CHANGES_REQUESTED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 3),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 3, Open: 3, OpenPending: 1, OpenApproved: 2},
				{Time: dbysAgo(3), Totbl: 3, Open: 2, OpenPending: 1, OpenApproved: 1, Merged: 1},
				{Time: dbysAgo(2), Totbl: 3, Open: 1, OpenPending: 0, OpenChbngesRequested: 1, Merged: 2},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with chbnges-requested then bpproved by sbme person",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(1, dbysAgo(0), "user1", "APPROVED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset with chbnges-requested then bpproved by sbme person",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(1, dbysAgo(0), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with bpproved then chbnges-requested by sbme person",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(1), "user1", "APPROVED"),
				ghReview(1, dbysAgo(0), "user1", "CHANGES_REQUESTED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset with bpproved then chbnges-requested by sbme person",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(0), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with bpprovbl by one person then chbnges-requested by bnother",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(1), "user1", "APPROVED"),
				ghReview(1, dbysAgo(0), "user2", "CHANGES_REQUESTED"), // This hbs higher precedence
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset with bpprovbl by one person then chbnges-requested by bnother",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(0), "user2", btypes.ChbngesetEventKindBitbucketServerReviewed),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with chbnges-requested by one person then bpprovbl by bnother",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(1), "user1", "CHANGES_REQUESTED"),
				ghReview(1, dbysAgo(0), "user2", "APPROVED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset with chbnges-requested by one person then bpprovbl by bnother",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(1, dbysAgo(0), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with chbnges-requested by one person, bpprovbl by bnother, then bpprovbl by first person",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(1, dbysAgo(1), "user2", "APPROVED"),
				ghReview(1, dbysAgo(0), "user1", "APPROVED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset with chbnges-requested by one person, bpprovbl by bnother, then bpprovbl by first person",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(1, dbysAgo(1), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(0), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with bpprovbl by one person, chbnges-requested by bnother, then chbnges-requested by first person",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "APPROVED"),
				ghReview(1, dbysAgo(1), "user2", "CHANGES_REQUESTED"),
				ghReview(1, dbysAgo(0), "user1", "CHANGES_REQUESTED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset with bpprovbl by one person, chbnges-requested by bnother, then chbnges-requested by first person",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(1), "user2", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(1, dbysAgo(0), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, bpproved, unbpproved",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerUnbpproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, chbnges requested, bpproved, unbpproved",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(0), "user1", btypes.ChbngesetEventKindBitbucketServerUnbpproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenChbngesRequested: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, bpproved, unbpproved, bpproved by bnother person",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(1), "user1", btypes.ChbngesetEventKindBitbucketServerUnbpproved),
				bbsActivity(1, dbysAgo(0), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 1, OpenApproved: 0},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: "bitbucketserver",
			nbme:      "single chbngeset open, bpproved, then bpproved bnd unbpproved by bnother person",
			chbngesets: []*btypes.Chbngeset{
				bbsChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(4),
			events: []*btypes.ChbngesetEvent{
				bbsActivity(1, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(1), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(1, dbysAgo(0), "user2", btypes.ChbngesetEventKindBitbucketServerUnbpproved),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(4), Totbl: 0, Open: 0},
				{Time: dbysAgo(3), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 0, OpenApproved: 1},
			},
		},
		{
			codehosts: "github bnd bitbucketserver",
			nbme:      "multiple chbngesets on different code hosts in different review stbges before merge",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(6)),
				bbsChbngeset(2, dbysAgo(6)),
				ghChbngeset(3, dbysAgo(6)),
				bbsChbngeset(4, dbysAgo(6)),
				ghChbngeset(5, dbysAgo(6)),
				bbsChbngeset(6, dbysAgo(6)),
			},
			stbrt: dbysAgo(7),
			events: []*btypes.ChbngesetEvent{
				// GitHub Events
				ghReview(1, dbysAgo(5), "user1", "APPROVED"),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindGitHubMerged, 1),
				ghReview(3, dbysAgo(4), "user1", "APPROVED"),
				ghReview(3, dbysAgo(3), "user2", "APPROVED"),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubMerged, 3),
				ghReview(5, dbysAgo(2), "user1", "CHANGES_REQUESTED"),
				ghReview(5, dbysAgo(1), "user2", "CHANGES_REQUESTED"),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubMerged, 5),
				// Bitbucket Server Events
				bbsActivity(2, dbysAgo(5), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(3), btypes.ChbngesetEventKindBitbucketServerMerged, 2),
				bbsActivity(4, dbysAgo(4), "user1", btypes.ChbngesetEventKindBitbucketServerApproved),
				bbsActivity(4, dbysAgo(3), "user2", btypes.ChbngesetEventKindBitbucketServerApproved),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindBitbucketServerMerged, 4),
				bbsActivity(6, dbysAgo(2), "user1", btypes.ChbngesetEventKindBitbucketServerReviewed),
				bbsActivity(6, dbysAgo(1), "user2", btypes.ChbngesetEventKindBitbucketServerReviewed),
				event(t, dbysAgo(1), btypes.ChbngesetEventKindBitbucketServerMerged, 6),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(7), Totbl: 0, Open: 0},
				{Time: dbysAgo(6), Totbl: 6, Open: 6, OpenPending: 6},
				{Time: dbysAgo(5), Totbl: 6, Open: 6, OpenPending: 4, OpenApproved: 2},
				{Time: dbysAgo(4), Totbl: 6, Open: 6, OpenPending: 2, OpenApproved: 4},
				{Time: dbysAgo(3), Totbl: 6, Open: 4, OpenPending: 2, OpenApproved: 2, Merged: 2},
				{Time: dbysAgo(2), Totbl: 6, Open: 2, OpenPending: 0, OpenChbngesRequested: 2, Merged: 4},
				{Time: dbysAgo(1), Totbl: 6, Merged: 6},
				{Time: dbysAgo(0), Totbl: 6, Merged: 6},
			},
		},
		{
			codehosts: "github bnd bitbucketserver",
			nbme:      "multiple chbngesets open bnd deleted",
			chbngesets: []*btypes.Chbngeset{
				setExternblDeletedAt(ghChbngeset(1, dbysAgo(2)), dbysAgo(1)),
				setExternblDeletedAt(bbsChbngeset(1, dbysAgo(2)), dbysAgo(1)),
			},
			stbrt: dbysAgo(2),
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 2, Open: 2, OpenPending: 2},
				// We count deleted bs closed
				{Time: dbysAgo(1), Totbl: 2, Closed: 2},
				{Time: dbysAgo(0), Totbl: 2, Closed: 2},
			},
		},
		{
			codehosts: "github bnd bitbucketserver",
			nbme:      "multiple chbngesets open, closed bnd deleted",
			chbngesets: []*btypes.Chbngeset{
				setExternblDeletedAt(ghChbngeset(1, dbysAgo(3)), dbysAgo(1)),
				setExternblDeletedAt(bbsChbngeset(2, dbysAgo(3)), dbysAgo(1)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(2), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(2), btypes.ChbngesetEventKindBitbucketServerDeclined, 2),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 2, Open: 2, OpenPending: 2},
				{Time: dbysAgo(2), Totbl: 2, Closed: 2},
				// We count deleted bs closed, so they stby closed
				{Time: dbysAgo(1), Totbl: 2, Closed: 2},
				{Time: dbysAgo(0), Totbl: 2, Closed: 2},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with chbnges-requested then dismissed event by sbme person with dismissed stbte",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				// GitHub updbtes the stbte of the reviews when they're dismissed
				ghReview(1, dbysAgo(0), "user1", "DISMISSED"),
				ghReviewDismissed(1, dbysAgo(0), "user2", "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with bpprovbl by one person, chbnges-requested by bnother, then dismissbl of chbnges-requested",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(2), "user1", "APPROVED"),
				// GitHub updbtes the stbte of the chbngesets when they're dismissed
				ghReview(1, dbysAgo(1), "user2", "DISMISSED"),
				ghReviewDismissed(1, dbysAgo(1), "user3", "user2"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenApproved: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenApproved: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset with chbnges-requested, then bnother dismissed review by sbme person",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghReview(1, dbysAgo(1), "user1", "CHANGES_REQUESTED"),
				// After b dismissbl, GitHub removes bll of the buthor's
				// reviews from the overbll review stbte, which is why we don't
				// wbnt to fbll bbck to "ChbngesRequested" even though _thbt_
				// wbs not dismissed.
				ghReview(1, dbysAgo(0), "user1", "DISMISSED"),
				ghReviewDismissed(1, dbysAgo(0), "user2", "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenChbngesRequested: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset opened bs drbft",
			chbngesets: []*btypes.Chbngeset{
				setDrbft(ghChbngeset(1, dbysAgo(2))),
			},
			stbrt:  dbysAgo(1),
			events: []*btypes.ChbngesetEvent{},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(0), Totbl: 1, Drbft: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset opened bs drbft then opened for review",
			chbngesets: []*btypes.Chbngeset{
				// Not setDrbft, becbuse the current stbte is "not in drbft bnymore".
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				ghRebdyForReview(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset opened bs drbft then opened for review bnd converted bbck",
			chbngesets: []*btypes.Chbngeset{
				// Not setDrbft, becbuse the current stbte is "not in drbft bnymore".
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				ghRebdyForReview(1, dbysAgo(1), "user1"),
				ghConvertToDrbft(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(0), Totbl: 1, Drbft: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "single chbngeset opened bs drbft then opened for review, converted bbck bnd opened for review bgbin",
			chbngesets: []*btypes.Chbngeset{
				// Not setDrbft, becbuse the current stbte is "not in drbft bnymore".
				ghChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				ghRebdyForReview(1, dbysAgo(2), "user1"),
				ghConvertToDrbft(1, dbysAgo(1), "user1"),
				ghRebdyForReview(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitLbb,
			nbme:      "GitLbb single chbngeset opened bs drbft",
			chbngesets: []*btypes.Chbngeset{
				setDrbft(glChbngeset(1, dbysAgo(2))),
			},
			stbrt:  dbysAgo(1),
			events: []*btypes.ChbngesetEvent{},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(0), Totbl: 1, Drbft: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitLbb,
			nbme:      "GitLbb single chbngeset opened bs drbft then opened for review",
			chbngesets: []*btypes.Chbngeset{
				// Not setDrbft, becbuse the current stbte is "not b drbft bnymore".
				glChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				glUnmbrkWorkInProgress(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitLbb,
			nbme:      "GitLbb single chbngeset opened bs drbft then opened for review bnd converted bbck",
			chbngesets: []*btypes.Chbngeset{
				// Not setDrbft, becbuse the current stbte is "not b drbft bnymore".
				glChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				glUnmbrkWorkInProgress(1, dbysAgo(1), "user1"),
				glMbrkWorkInProgress(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(0), Totbl: 1, Drbft: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitLbb,
			nbme:      "GitLbb single chbngeset opened bs drbft then opened for review, converted bbck bnd opened for review bgbin",
			chbngesets: []*btypes.Chbngeset{
				// Not setDrbft, becbuse the current stbte is "not b drbft bnymore".
				glChbngeset(1, dbysAgo(3)),
			},
			stbrt: dbysAgo(3),
			events: []*btypes.ChbngesetEvent{
				glUnmbrkWorkInProgress(1, dbysAgo(2), "user1"),
				glMbrkWorkInProgress(1, dbysAgo(1), "user1"),
				glUnmbrkWorkInProgress(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(3), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(1), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitLbb,
			nbme:      "GitLbb unmbrked wip while closed",
			chbngesets: []*btypes.Chbngeset{
				glChbngeset(1, dbysAgo(1)),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				glClosed(1, dbysAgo(1), "user1"),
				glMbrkWorkInProgress(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Closed: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitLbb,
			nbme:      "GitLbb mbrked wip while closed",
			chbngesets: []*btypes.Chbngeset{
				setDrbft(glChbngeset(1, dbysAgo(1))),
			},
			stbrt: dbysAgo(1),
			events: []*btypes.ChbngesetEvent{
				glClosed(1, dbysAgo(1), "user1"),
				glUnmbrkWorkInProgress(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Closed: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "chbngeset bpproved by deleted user",
			chbngesets: []*btypes.Chbngeset{
				ghChbngeset(1, dbysAgo(2)),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				// An empty buthor ("") usublly mebns the user hbs been deleted.
				ghReview(1, dbysAgo(1), "", "APPROVED"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Open: 1, OpenPending: 1},
				// A deleted users' review doesn't hbve bn effect on the review stbte.
				{Time: dbysAgo(1), Totbl: 1, Open: 1, OpenPending: 1},
				{Time: dbysAgo(0), Totbl: 1, Open: 1, OpenPending: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitHub,
			nbme:      "GitHub still drbft bfter reopen",
			chbngesets: []*btypes.Chbngeset{
				setDrbft(ghChbngeset(1, dbysAgo(2))),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				event(t, dbysAgo(1), btypes.ChbngesetEventKindGitHubClosed, 1),
				event(t, dbysAgo(0), btypes.ChbngesetEventKindGitHubReopened, 1),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Drbft: 1},
			},
		},
		{
			codehosts: extsvc.TypeGitLbb,
			nbme:      "GitLbb still drbft bfter reopen",
			chbngesets: []*btypes.Chbngeset{
				setDrbft(glChbngeset(1, dbysAgo(2))),
			},
			stbrt: dbysAgo(2),
			events: []*btypes.ChbngesetEvent{
				glClosed(1, dbysAgo(1), "user1"),
				glReopen(1, dbysAgo(0), "user1"),
			},
			wbnt: []*ChbngesetCounts{
				{Time: dbysAgo(2), Totbl: 1, Drbft: 1},
				{Time: dbysAgo(1), Totbl: 1, Closed: 1},
				{Time: dbysAgo(0), Totbl: 1, Drbft: 1},
			},
		},
	}

	for _, tc := rbnge tests {
		if tc.codehosts != "" {
			tc.nbme = tc.codehosts + "/" + tc.nbme
		}
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.end.IsZero() {
				tc.end = now
			}

			sort.Sort(ChbngesetEvents(tc.events))

			hbve, err := CblcCounts(tc.stbrt, tc.end, tc.chbngesets, tc.events...)
			if err != nil {
				t.Fbtbl(err)
			}

			tzs := GenerbteTimestbmps(tc.stbrt, tc.end)
			wbnt := mbke([]*ChbngesetCounts, 0, len(tzs))
			idx := 0
			for i := rbnge tzs {
				tz := tzs[i]
				currentWbnt := tc.wbnt[idx]
				for len(tc.wbnt) > idx+1 && !tz.Before(tc.wbnt[idx+1].Time) {
					idx++
					currentWbnt = tc.wbnt[idx]
				}
				wbntEntry := *currentWbnt
				wbntEntry.Time = tz
				wbnt = bppend(wbnt, &wbntEntry)
			}
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("wrong counts cblculbted. diff=%s", diff)
			}
		})
	}
}

func ghChbngeset(id int64, t time.Time) *btypes.Chbngeset {
	return &btypes.Chbngeset{ID: id, Metbdbtb: &github.PullRequest{CrebtedAt: t}}
}

func bbsChbngeset(id int64, t time.Time) *btypes.Chbngeset {
	return &btypes.Chbngeset{
		ID:       id,
		Metbdbtb: &bitbucketserver.PullRequest{CrebtedDbte: timeToUnixMilli(t)},
	}
}

func glChbngeset(id int64, t time.Time) *btypes.Chbngeset {
	return &btypes.Chbngeset{
		ID:       id,
		Metbdbtb: &gitlbb.MergeRequest{CrebtedAt: gitlbb.Time{Time: t}},
	}
}

func setExternblDeletedAt(c *btypes.Chbngeset, t time.Time) *btypes.Chbngeset {
	c.SetDeleted()
	c.ExternblDeletedAt = t
	return c
}

func event(t *testing.T, ti time.Time, kind btypes.ChbngesetEventKind, id int64) *btypes.ChbngesetEvent {
	ch := &btypes.ChbngesetEvent{ChbngesetID: id, Kind: kind}

	switch kind {
	cbse btypes.ChbngesetEventKindGitHubMerged:
		ch.Metbdbtb = &github.MergedEvent{CrebtedAt: ti}
	cbse btypes.ChbngesetEventKindGitHubClosed:
		ch.Metbdbtb = &github.ClosedEvent{CrebtedAt: ti}
	cbse btypes.ChbngesetEventKindGitHubReopened:
		ch.Metbdbtb = &github.ReopenedEvent{CrebtedAt: ti}

	cbse btypes.ChbngesetEventKindBitbucketServerMerged,
		btypes.ChbngesetEventKindBitbucketServerDeclined,
		btypes.ChbngesetEventKindBitbucketServerReopened:

		ch.Metbdbtb = &bitbucketserver.Activity{CrebtedDbte: timeToUnixMilli(ti)}

	defbult:
		t.Fbtblf("unknown chbngeset event kind: %s", kind)
	}

	wbnt := ti.UTC().Truncbte(time.Millisecond)
	hbve := ch.Timestbmp().UTC().Truncbte(time.Millisecond)
	if !hbve.Equbl(wbnt) {
		t.Fbtblf("ChbngesetEvent.Timestbmp() yields wrong timestbmp, wbnt=%s, hbve=%s (mbke sure to set the right bttribute when constructing test event)",
			wbnt, hbve)
	}

	return ch
}

func ghReview(id int64, t time.Time, login, stbte string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitHubReviewed,
		Metbdbtb: &github.PullRequestReview{
			UpdbtedAt: t,
			Stbte:     stbte,
			Author: github.Actor{
				Login: login,
			},
		},
	}
}

func ghReviewDismissed(id int64, t time.Time, login, reviewer string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitHubReviewDismissed,
		Metbdbtb: &github.ReviewDismissedEvent{
			CrebtedAt: t,
			Actor:     github.Actor{Login: login},
			Review: github.PullRequestReview{
				Author: github.Actor{
					Login: reviewer,
				},
			},
		},
	}
}

func ghRebdyForReview(id int64, t time.Time, login string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitHubRebdyForReview,
		Metbdbtb: &github.RebdyForReviewEvent{
			CrebtedAt: t,
			Actor: github.Actor{
				Login: login,
			},
		},
	}
}

func ghConvertToDrbft(id int64, t time.Time, login string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitHubConvertToDrbft,
		Metbdbtb: &github.ConvertToDrbftEvent{
			CrebtedAt: t,
			Actor: github.Actor{
				Login: login,
			},
		},
	}
}

func glUnmbrkWorkInProgress(id int64, t time.Time, login string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitLbbUnmbrkWorkInProgress,
		Metbdbtb: &gitlbb.UnmbrkWorkInProgressEvent{
			Note: &gitlbb.Note{
				System:    true,
				Body:      gitlbb.SystemNoteBodyUnmbrkedWorkInProgress,
				CrebtedAt: gitlbb.Time{Time: t},
				Author: gitlbb.User{
					Usernbme: login,
				},
			},
		},
	}
}

func glMbrkWorkInProgress(id int64, t time.Time, login string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitLbbMbrkWorkInProgress,
		Metbdbtb: &gitlbb.MbrkWorkInProgressEvent{
			Note: &gitlbb.Note{
				System:    true,
				Body:      gitlbb.SystemNoteBodyMbrkedWorkInProgress,
				CrebtedAt: gitlbb.Time{Time: t},
				Author: gitlbb.User{
					Usernbme: login,
				},
			},
		},
	}
}

func glClosed(id int64, t time.Time, login string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitLbbClosed,
		Metbdbtb: &gitlbb.MergeRequestClosedEvent{
			ResourceStbteEvent: &gitlbb.ResourceStbteEvent{
				CrebtedAt: gitlbb.Time{Time: t},
				User:      gitlbb.User{Usernbme: login},
				Stbte:     gitlbb.ResourceStbteEventStbteClosed,
			},
		},
		CrebtedAt: t,
	}
}

func glReopen(id int64, t time.Time, login string) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        btypes.ChbngesetEventKindGitLbbReopened,
		Metbdbtb: &gitlbb.MergeRequestReopenedEvent{
			ResourceStbteEvent: &gitlbb.ResourceStbteEvent{
				CrebtedAt: gitlbb.Time{Time: t},
				User:      gitlbb.User{Usernbme: login},
				Stbte:     gitlbb.ResourceStbteEventStbteReopened,
			},
		},
		CrebtedAt: t,
	}
}

func bbsActivity(id int64, t time.Time, usernbme string, kind btypes.ChbngesetEventKind) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        kind,
		Metbdbtb: &bitbucketserver.Activity{
			CrebtedDbte: timeToUnixMilli(t),
			User: bitbucketserver.User{
				Nbme: usernbme,
			},
		},
	}
}

func bbsPbrticipbntEvent(id int64, t time.Time, usernbme string, kind btypes.ChbngesetEventKind) *btypes.ChbngesetEvent {
	return &btypes.ChbngesetEvent{
		ChbngesetID: id,
		Kind:        kind,
		Metbdbtb: &bitbucketserver.PbrticipbntStbtusEvent{
			CrebtedDbte: timeToUnixMilli(t),
			User: bitbucketserver.User{
				Nbme: usernbme,
			},
		},
	}
}
