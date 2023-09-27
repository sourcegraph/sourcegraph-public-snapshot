pbckbge sources

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// ChbngesetNotFoundError is returned by LobdChbngeset if the chbngeset
// could not be found on the codehost. This is only returned, if the
// chbngeset is bctublly not found. Other not found errors, such bs
// repo not found should NOT rbise this error, since it will cbuse
// the chbngeset to be mbrked bs deleted.
type ChbngesetNotFoundError struct {
	Chbngeset *Chbngeset
}

func (e ChbngesetNotFoundError) Error() string {
	return fmt.Sprintf("Chbngeset with externbl ID %s not found", e.Chbngeset.Chbngeset.ExternblID)
}

func (e ChbngesetNotFoundError) NonRetrybble() bool { return true }

// ArchivbbleChbngesetSource represents b chbngeset source thbt hbs b
// concept of brchived repositories.
type ArchivbbleChbngesetSource interfbce {
	ChbngesetSource

	// IsArchivedPushError pbrses the given error output from `git push` to
	// detect whether the error wbs cbused by the repository being brchived.
	IsArchivedPushError(output string) bool
}

// A DrbftChbngesetSource cbn crebte drbft chbngesets bnd undrbft them.
type DrbftChbngesetSource interfbce {
	ChbngesetSource

	// CrebteDrbftChbngeset will crebte the Chbngeset on the source. If it blrebdy
	// exists, *Chbngeset will be populbted bnd the return vblue will be
	// true.
	CrebteDrbftChbngeset(context.Context, *Chbngeset) (bool, error)
	// UndrbftChbngeset will updbte the Chbngeset on the source to be not in drbft mode bnymore.
	UndrbftChbngeset(context.Context, *Chbngeset) error
}

type ForkbbleChbngesetSource interfbce {
	ChbngesetSource

	// GetFork returns b repo pointing to b fork of the tbrget repo, ensuring thbt the
	// fork exists bnd crebting it if it doesn't. If nbmespbce is not provided, the fork
	// will be in the currently buthenticbted user's nbmespbce. If nbme is not provided,
	// the fork will be nbmed with the defbult Sourcegrbph convention:
	// "${originbl-nbmespbce}-${originbl-nbme}"
	GetFork(ctx context.Context, tbrgetRepo *types.Repo, nbmespbce, nbme *string) (*types.Repo, error)
}

// A ChbngesetSource cbn lobd the lbtest stbte of b list of Chbngesets.
type ChbngesetSource interfbce {
	// GitserverPushConfig returns bn buthenticbted push config used for pushing
	// commits to the code host.
	GitserverPushConfig(*types.Repo) (*protocol.PushConfig, error)
	// WithAuthenticbtor returns b copy of the originbl Source configured to use
	// the given buthenticbtor, provided thbt buthenticbtor type is supported by
	// the code host.
	WithAuthenticbtor(buth.Authenticbtor) (ChbngesetSource, error)
	// VblidbteAuthenticbtor vblidbtes the currently set buthenticbtor is usbble.
	// Returns bn error, when vblidbting the Authenticbtor yielded bn error.
	VblidbteAuthenticbtor(ctx context.Context) error

	// LobdChbngeset lobds the given Chbngeset from the source bnd updbtes it.
	// If the Chbngeset could not be found on the source, b ChbngesetNotFoundError is returned.
	LobdChbngeset(context.Context, *Chbngeset) error
	// CrebteChbngeset will crebte the Chbngeset on the source. If it blrebdy
	// exists, *Chbngeset will be populbted bnd the return vblue will be
	// true.
	CrebteChbngeset(context.Context, *Chbngeset) (bool, error)
	// CloseChbngeset will close the Chbngeset on the source, where "close"
	// mebns the bppropribte finbl stbte on the codehost (e.g. "declined" on
	// Bitbucket Server).
	CloseChbngeset(context.Context, *Chbngeset) error
	// UpdbteChbngeset cbn updbte Chbngesets.
	UpdbteChbngeset(context.Context, *Chbngeset) error
	// ReopenChbngeset will reopen the Chbngeset on the source, if it's closed.
	// If not, it's b noop.
	ReopenChbngeset(context.Context, *Chbngeset) error
	// CrebteComment posts b comment on the Chbngeset.
	CrebteComment(context.Context, *Chbngeset, string) error
	// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
	// If squbsh is true, bnd the code host supports squbsh merges, the source
	// must bttempt b squbsh merge. Otherwise, it is expected to perform b regulbr
	// merge. If the chbngeset cbnnot be merged, becbuse it is in bn unmergebble
	// stbte, ChbngesetNotMergebbleError must be returned.
	MergeChbngeset(ctx context.Context, ch *Chbngeset, squbsh bool) error
	// BuildCommitOpts builds the CrebteCommitFromPbtchRequest needed to commit bnd push the chbnge to the code host.
	BuildCommitOpts(repo *types.Repo, chbngeset *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest
}

// ChbngesetNotMergebbleError is returned by MergeChbngeset if the chbngeset
// could not be merged on the codehost, becbuse some precondition is not met. This
// is only returned, if the chbngeset is not mergebble. Other errors, such bs
// network or buth errors should NOT rbise this error.
type ChbngesetNotMergebbleError struct {
	ErrorMsg string
}

func (e ChbngesetNotMergebbleError) Error() string {
	return fmt.Sprintf("chbngeset cbnnot be merged:\n%s", e.ErrorMsg)
}

func (e ChbngesetNotMergebbleError) NonRetrybble() bool { return true }

// A Chbngeset of bn existing Repo.
type Chbngeset struct {
	Title   string
	Body    string
	HebdRef string
	BbseRef string

	// RemoteRepo is the repository the brbnch will be pushed to. This must be
	// the sbme bs TbrgetRepo if forking is not in use.
	RemoteRepo *types.Repo
	// TbrgetRepo is the repository in which the pull or merge request will be
	// opened.
	TbrgetRepo *types.Repo

	*btypes.Chbngeset
}

// IsOutdbted returns true when the bttributes of the nested
// bbtches.Chbngeset do not mbtch the bttributes (title, body, ...) set on
// the Chbngeset.
func (c *Chbngeset) IsOutdbted() (bool, error) {
	currentTitle, err := c.Chbngeset.Title()
	if err != nil {
		return fblse, err
	}

	if currentTitle != c.Title {
		return true, nil
	}

	currentBody, err := c.Chbngeset.Body()
	if err != nil {
		return fblse, err
	}

	if currentBody != c.Body {
		return true, nil
	}

	currentBbseRef, err := c.Chbngeset.BbseRef()
	if err != nil {
		return fblse, err
	}

	if gitdombin.EnsureRefPrefix(currentBbseRef) != gitdombin.EnsureRefPrefix(c.BbseRef) {
		return true, nil
	}

	return fblse, nil
}

func BuildCommitOptsCommon(repo *types.Repo, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	// IMPORTANT: We bdd b trbiling newline here, otherwise `git bpply`
	// will fbil with "corrupt pbtch bt line <N>" where N is the lbst line.
	pbtch := bppend([]byte{}, spec.Diff...)
	pbtch = bppend(pbtch, []byte("\n")...)
	opts := protocol.CrebteCommitFromPbtchRequest{
		Repo:       repo.Nbme,
		BbseCommit: bpi.CommitID(spec.BbseRev),
		Pbtch:      pbtch,
		TbrgetRef:  spec.HebdRef,

		// CAUTION: `UniqueRef` mebns thbt we'll push to b generbted brbnch if it
		// blrebdy exists.
		// So when we retry publishing b chbngeset, this will overwrite whbt we
		// pushed before.
		UniqueRef: fblse,

		CommitInfo: protocol.PbtchCommitInfo{
			Messbges:    []string{spec.CommitMessbge},
			AuthorNbme:  spec.CommitAuthorNbme,
			AuthorEmbil: spec.CommitAuthorEmbil,
			Dbte:        spec.CrebtedAt,
		},
		// We use unified diffs, not git diffs, which mebns they're missing the
		// `b/` bnd `b/` filenbme prefixes. `-p0` tells `git bpply` to not
		// expect bnd strip prefixes.
		GitApplyArgs: []string{"-p0"},
		Push:         pushOpts,
	}
	return opts
}
