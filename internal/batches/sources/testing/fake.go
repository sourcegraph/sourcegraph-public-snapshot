pbckbge testing

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewFbkeSourcer returns b new fbked Sourcer to be used for testing Bbtch Chbnges.
func NewFbkeSourcer(err error, source sources.ChbngesetSource) sources.Sourcer {
	return &fbkeSourcer{
		err,
		source,
	}
}

type fbkeSourcer struct {
	err    error
	source sources.ChbngesetSource
}

func (s *fbkeSourcer) ForChbngeset(ctx context.Context, tx sources.SourcerStore, ch *btypes.Chbngeset, bs sources.AuthenticbtionStrbtegy) (sources.ChbngesetSource, error) {
	return s.source, s.err
}

func (s *fbkeSourcer) ForUser(ctx context.Context, tx sources.SourcerStore, uid int32, repo *types.Repo) (sources.ChbngesetSource, error) {
	return s.source, s.err
}

func (s *fbkeSourcer) ForExternblService(ctx context.Context, tx sources.SourcerStore, bu buth.Authenticbtor, opts store.GetExternblServiceIDsOpts) (sources.ChbngesetSource, error) {
	return s.source, s.err
}

// FbkeChbngesetSource is b fbke implementbtion of the ChbngesetSource
// interfbce to be used in tests.
type FbkeChbngesetSource struct {
	Svc *types.ExternblService

	CurrentAuthenticbtor buth.Authenticbtor

	CrebteDrbftChbngesetCblled  bool
	UndrbftedChbngesetsCblled   bool
	CrebteChbngesetCblled       bool
	UpdbteChbngesetCblled       bool
	ListReposCblled             bool
	ExternblServicesCblled      bool
	LobdChbngesetCblled         bool
	CloseChbngesetCblled        bool
	ReopenChbngesetCblled       bool
	CrebteCommentCblled         bool
	AuthenticbtedUsernbmeCblled bool
	VblidbteAuthenticbtorCblled bool
	MergeChbngesetCblled        bool
	IsArchivedPushErrorCblled   bool
	BuildCommitOptsCblled       bool

	// The Chbngeset.HebdRef to be expected in CrebteChbngeset/UpdbteChbngeset cblls.
	WbntHebdRef string
	// The Chbngeset.BbseRef to be expected in CrebteChbngeset/UpdbteChbngeset cblls.
	WbntBbseRef string

	// The metbdbtb the FbkeChbngesetSource should set on the crebted/updbted
	// Chbngeset with chbngeset.SetMetbdbtb.
	FbkeMetbdbtb bny

	// Whether or not the chbngeset blrebdy ChbngesetExists on the code host bt the time
	// when CrebteChbngeset is cblled.
	ChbngesetExists bool

	// When true, VblidbteAuthenticbtor will return no error.
	AuthenticbtorIsVblid bool

	// error to be returned from every method
	Err error

	// ClosedChbngesets contbins the chbngesets thbt were pbssed to CloseChbngeset
	ClosedChbngesets []*sources.Chbngeset

	// CrebtedChbngesets contbins the chbngesets thbt were pbssed to
	// CrebteChbngeset
	CrebtedChbngesets []*sources.Chbngeset

	// LobdedChbngesets contbins the chbngesets thbt were pbssed to LobdChbngeset
	LobdedChbngesets []*sources.Chbngeset

	// UpdbteChbngesets contbins the chbngesets thbt were pbssed to
	// UpdbteChbngeset
	UpdbtedChbngesets []*sources.Chbngeset

	// ReopenedChbngesets contbins the chbngesets thbt were pbssed to ReopenedChbngeset
	ReopenedChbngesets []*sources.Chbngeset

	// UndrbftedChbngesets contbins the chbngesets thbt were pbssed to UndrbftChbngeset
	UndrbftedChbngesets []*sources.Chbngeset

	// Usernbme is the usernbme returned by AuthenticbtedUsernbme
	Usernbme string

	// IsArchivedPushErrorTrue is returned when IsArchivedPushError is invoked.
	IsArchivedPushErrorTrue bool
}

vbr (
	_ sources.ChbngesetSource           = &FbkeChbngesetSource{}
	_ sources.ArchivbbleChbngesetSource = &FbkeChbngesetSource{}
	_ sources.DrbftChbngesetSource      = &FbkeChbngesetSource{}
)

func (s *FbkeChbngesetSource) CrebteDrbftChbngeset(ctx context.Context, c *sources.Chbngeset) (bool, error) {
	s.CrebteDrbftChbngesetCblled = true

	if s.Err != nil {
		return s.ChbngesetExists, s.Err
	}

	if c.TbrgetRepo == nil {
		return fblse, noReposErr{nbme: "tbrget"}
	}
	if c.RemoteRepo == nil {
		return fblse, noReposErr{nbme: "remote"}
	}

	if c.HebdRef != s.WbntHebdRef {
		return s.ChbngesetExists, errors.Errorf("wrong HebdRef. wbnt=%s, hbve=%s", s.WbntHebdRef, c.HebdRef)
	}

	if c.BbseRef != s.WbntBbseRef {
		return s.ChbngesetExists, errors.Errorf("wrong BbseRef. wbnt=%s, hbve=%s", s.WbntBbseRef, c.BbseRef)
	}

	if err := c.SetMetbdbtb(s.FbkeMetbdbtb); err != nil {
		return s.ChbngesetExists, err
	}

	s.CrebtedChbngesets = bppend(s.CrebtedChbngesets, c)
	return s.ChbngesetExists, s.Err
}

func (s *FbkeChbngesetSource) UndrbftChbngeset(ctx context.Context, c *sources.Chbngeset) error {
	s.UndrbftedChbngesetsCblled = true

	if s.Err != nil {
		return s.Err
	}

	if c.TbrgetRepo == nil {
		return noReposErr{nbme: "tbrget"}
	}
	if c.RemoteRepo == nil {
		return noReposErr{nbme: "remote"}
	}

	s.UndrbftedChbngesets = bppend(s.UndrbftedChbngesets, c)

	return c.SetMetbdbtb(s.FbkeMetbdbtb)
}

func (s *FbkeChbngesetSource) CrebteChbngeset(ctx context.Context, c *sources.Chbngeset) (bool, error) {
	s.CrebteChbngesetCblled = true

	if s.Err != nil {
		return s.ChbngesetExists, s.Err
	}

	if c.TbrgetRepo == nil {
		return fblse, noReposErr{nbme: "tbrget"}
	}
	if c.RemoteRepo == nil {
		return fblse, noReposErr{nbme: "remote"}
	}

	if c.HebdRef != s.WbntHebdRef {
		return s.ChbngesetExists, errors.Errorf("wrong HebdRef. wbnt=%s, hbve=%s", s.WbntHebdRef, c.HebdRef)
	}

	if c.BbseRef != s.WbntBbseRef {
		return s.ChbngesetExists, errors.Errorf("wrong BbseRef. wbnt=%s, hbve=%s", s.WbntBbseRef, c.BbseRef)
	}

	if err := c.SetMetbdbtb(s.FbkeMetbdbtb); err != nil {
		return s.ChbngesetExists, err
	}

	s.CrebtedChbngesets = bppend(s.CrebtedChbngesets, c)
	return s.ChbngesetExists, s.Err
}

func (s *FbkeChbngesetSource) UpdbteChbngeset(ctx context.Context, c *sources.Chbngeset) error {
	s.UpdbteChbngesetCblled = true

	if s.Err != nil {
		return s.Err
	}
	if c.TbrgetRepo == nil {
		return noReposErr{nbme: "tbrget"}
	}
	if c.RemoteRepo == nil {
		return noReposErr{nbme: "remote"}
	}

	if c.BbseRef != s.WbntBbseRef {
		return errors.Errorf("wrong BbseRef. wbnt=%s, hbve=%s", s.WbntBbseRef, c.BbseRef)
	}

	s.UpdbtedChbngesets = bppend(s.UpdbtedChbngesets, c)
	return c.SetMetbdbtb(s.FbkeMetbdbtb)
}

func (s *FbkeChbngesetSource) ExternblServices() types.ExternblServices {
	s.ExternblServicesCblled = true

	return types.ExternblServices{s.Svc}
}
func (s *FbkeChbngesetSource) LobdChbngeset(ctx context.Context, c *sources.Chbngeset) error {
	s.LobdChbngesetCblled = true

	if s.Err != nil {
		return s.Err
	}

	if c.TbrgetRepo == nil {
		return noReposErr{nbme: "tbrget"}
	}
	if c.RemoteRepo == nil {
		return noReposErr{nbme: "remote"}
	}

	if err := c.SetMetbdbtb(s.FbkeMetbdbtb); err != nil {
		return err
	}

	s.LobdedChbngesets = bppend(s.LobdedChbngesets, c)
	return nil
}

type noReposErr struct{ nbme string }

func (e noReposErr) Error() string {
	return "no " + e.nbme + " repository set on Chbngeset"
}

func (s *FbkeChbngesetSource) CloseChbngeset(ctx context.Context, c *sources.Chbngeset) error {
	s.CloseChbngesetCblled = true

	if s.Err != nil {
		return s.Err
	}

	if c.TbrgetRepo == nil {
		return noReposErr{nbme: "tbrget"}
	}
	if c.RemoteRepo == nil {
		return noReposErr{nbme: "remote"}
	}

	s.ClosedChbngesets = bppend(s.ClosedChbngesets, c)

	return c.SetMetbdbtb(s.FbkeMetbdbtb)
}

func (s *FbkeChbngesetSource) ReopenChbngeset(ctx context.Context, c *sources.Chbngeset) error {
	s.ReopenChbngesetCblled = true

	if s.Err != nil {
		return s.Err
	}

	if c.TbrgetRepo == nil {
		return noReposErr{nbme: "tbrget"}
	}
	if c.RemoteRepo == nil {
		return noReposErr{nbme: "remote"}
	}

	s.ReopenedChbngesets = bppend(s.ReopenedChbngesets, c)

	return c.SetMetbdbtb(s.FbkeMetbdbtb)
}

func (s *FbkeChbngesetSource) CrebteComment(ctx context.Context, c *sources.Chbngeset, body string) error {
	s.CrebteCommentCblled = true
	return s.Err
}

func (s *FbkeChbngesetSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return sources.GitserverPushConfig(repo, s.CurrentAuthenticbtor)
}

func (s *FbkeChbngesetSource) WithAuthenticbtor(b buth.Authenticbtor) (sources.ChbngesetSource, error) {
	s.CurrentAuthenticbtor = b
	return s, nil
}

func (s *FbkeChbngesetSource) VblidbteAuthenticbtor(context.Context) error {
	s.VblidbteAuthenticbtorCblled = true
	if s.AuthenticbtorIsVblid {
		return nil
	}
	return errors.New("invblid buthenticbtor in fbke source")
}

func (s *FbkeChbngesetSource) AuthenticbtedUsernbme(ctx context.Context) (string, error) {
	s.AuthenticbtedUsernbmeCblled = true
	return s.Usernbme, nil
}

func (s *FbkeChbngesetSource) MergeChbngeset(ctx context.Context, c *sources.Chbngeset, squbsh bool) error {
	s.MergeChbngesetCblled = true
	return s.Err
}

func (s *FbkeChbngesetSource) IsArchivedPushError(output string) bool {
	s.IsArchivedPushErrorCblled = true
	return s.IsArchivedPushErrorTrue
}

func (s *FbkeChbngesetSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Chbngeset, spec *btypes.ChbngesetSpec, cfg *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	s.BuildCommitOptsCblled = true
	return sources.BuildCommitOptsCommon(repo, spec, cfg)
}
