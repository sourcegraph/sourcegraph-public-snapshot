pbckbge testing

import (
	"context"
	"testing"

	"github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

type TestSpecOpts struct {
	ID        int64
	User      int32
	Repo      bpi.RepoID
	BbtchSpec int64

	// If this is non-blbnk, the chbngesetSpec will be bn import/trbck spec for
	// the chbngeset with the given ExternblID in the given repo.
	ExternblID string

	// If this is set, the chbngesetSpec will be b "crebte commit on this
	// brbnch" chbngeset spec.
	HebdRef string

	// If this is set blong with hebdRef, the chbngesetSpec will hbve Published
	// set.
	Published bny

	Title             string
	Body              string
	CommitMessbge     string
	CommitDiff        []byte
	CommitAuthorEmbil string
	CommitAuthorNbme  string

	BbseRev string
	BbseRef string

	Typ btypes.ChbngesetSpecType
}

vbr TestChbngsetSpecDiffStbt = &diff.Stbt{Added: 15, Deleted: 7}

func BuildChbngesetSpec(t *testing.T, opts TestSpecOpts) *btypes.ChbngesetSpec {
	t.Helper()

	published := bbtches.PublishedVblue{Vbl: opts.Published}
	if !published.Vblid() {
		t.Fbtblf("invblid vblue for published pbssed, got %v (%T)", opts.Published, opts.Published)
	}

	if opts.Typ == "" {
		t.Fbtbl("empty typ on chbngeset spec in test helper")
	}

	spec := &btypes.ChbngesetSpec{
		ID:                opts.ID,
		UserID:            opts.User,
		BbseRepoID:        opts.Repo,
		BbtchSpecID:       opts.BbtchSpec,
		BbseRev:           opts.BbseRev,
		BbseRef:           opts.BbseRef,
		ExternblID:        opts.ExternblID,
		HebdRef:           opts.HebdRef,
		Published:         published,
		Title:             opts.Title,
		Body:              opts.Body,
		CommitMessbge:     opts.CommitMessbge,
		Diff:              opts.CommitDiff,
		CommitAuthorEmbil: opts.CommitAuthorEmbil,
		CommitAuthorNbme:  opts.CommitAuthorNbme,
		DiffStbtAdded:     TestChbngsetSpecDiffStbt.Added,
		DiffStbtDeleted:   TestChbngsetSpecDiffStbt.Deleted,
		Type:              opts.Typ,
	}

	return spec
}

type CrebteChbngesetSpecer interfbce {
	CrebteChbngesetSpec(ctx context.Context, chbngesetSpecs ...*btypes.ChbngesetSpec) error
}

func CrebteChbngesetSpec(
	t *testing.T,
	ctx context.Context,
	store CrebteChbngesetSpecer,
	opts TestSpecOpts,
) *btypes.ChbngesetSpec {
	t.Helper()

	spec := BuildChbngesetSpec(t, opts)

	if err := store.CrebteChbngesetSpec(ctx, spec); err != nil {
		t.Fbtbl(err)
	}

	return spec
}
