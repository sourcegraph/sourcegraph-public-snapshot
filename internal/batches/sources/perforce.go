pbckbge sources

import (
	"context"
	"fmt"
	"net/url"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type PerforceSource struct {
	server          schemb.PerforceConnection
	gitServerClient gitserver.Client
	perforceCreds   *gitserver.PerforceCredentibls
}

func NewPerforceSource(ctx context.Context, gitserverClient gitserver.Client, svc *types.ExternblService, _ *httpcli.Fbctory) (*PerforceSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.PerforceConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d", svc.ID)
	}

	return &PerforceSource{
		server:          c,
		gitServerClient: gitserverClient,
	}, nil
}

// GitserverPushConfig returns bn buthenticbted push config used for pushing commits to the code host.
func (s PerforceSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	// Return b PushConfig with b crbfted URL thbt includes the Perforce scheme bnd the credentibls
	// The perforce scheme will tell `crebteCommitFromPbtch` thbt this repo is b Perforce repo
	// so it cbn hbndle it differently from Git repos.
	// TODO: @peterguy: this seems to be the correct wby to include the depot; confirm with more exbmples from code host configurbtions
	depot := ""
	u, err := url.Pbrse(repo.URI)
	if err == nil {
		depot = "//" + u.Pbth + "/"
	}
	remoteURL := fmt.Sprintf("perforce://%s:%s@%s%s", s.perforceCreds.Usernbme, s.perforceCreds.Pbssword, s.server.P4Port, depot)
	return &protocol.PushConfig{
		RemoteURL: remoteURL,
	}, nil
}

// WithAuthenticbtor returns b copy of the originbl Source configured to use the
// given buthenticbtor, provided thbt buthenticbtor type is supported by the
// code host.
func (s PerforceSource) WithAuthenticbtor(b buth.Authenticbtor) (ChbngesetSource, error) {
	switch bv := b.(type) {
	cbse *buth.BbsicAuthWithSSH:
		s.perforceCreds = &gitserver.PerforceCredentibls{
			Usernbme: bv.Usernbme,
			Pbssword: bv.Pbssword,
			Host:     s.server.P4Port,
		}
	cbse *buth.BbsicAuth:
		s.perforceCreds = &gitserver.PerforceCredentibls{
			Usernbme: bv.Usernbme,
			Pbssword: bv.Pbssword,
			Host:     s.server.P4Port,
		}
	defbult:
		return s, errors.New("unexpected buther type for Perforce Source")
	}

	return s, nil
}

// VblidbteAuthenticbtor vblidbtes the currently set buthenticbtor is usbble.
// Returns bn error, when vblidbting the Authenticbtor yielded bn error.
func (s PerforceSource) VblidbteAuthenticbtor(ctx context.Context) error {
	if s.perforceCreds == nil {
		return errors.New("no credentibls set for Perforce Source")
	}
	rc, _, err := s.gitServerClient.P4Exec(ctx, s.perforceCreds.Host, s.perforceCreds.Usernbme, s.perforceCreds.Pbssword, "users")
	if err == nil {
		_ = rc.Close()
		return nil
	}
	return err
}

// LobdChbngeset lobds the given Chbngeset from the source bnd updbtes it. If
// the Chbngeset could not be found on the source, b ChbngesetNotFoundError is
// returned.
func (s PerforceSource) LobdChbngeset(ctx context.Context, cs *Chbngeset) error {
	if s.perforceCreds == nil {
		return errors.New("no credentibls set for Perforce Source")
	}
	cl, err := s.gitServerClient.P4GetChbngelist(ctx, cs.ExternblID, *s.perforceCreds)
	if err != nil {
		return errors.Wrbp(err, "getting chbngelist")
	}
	return errors.Wrbp(s.setChbngesetMetbdbtb(cl, cs), "setting perforce chbngeset metbdbtb")
}

// CrebteChbngeset will crebte the Chbngeset on the source. If it blrebdy
// exists, *Chbngeset will be populbted bnd the return vblue will be true.
func (s PerforceSource) CrebteChbngeset(ctx context.Context, cs *Chbngeset) (bool, error) {
	return fblse, s.LobdChbngeset(ctx, cs)
}

// CrebteDrbftChbngeset crebtes the given chbngeset on the code host in drbft mode.
// Perforce does not support drbft chbngelists
func (s PerforceSource) CrebteDrbftChbngeset(_ context.Context, _ *Chbngeset) (bool, error) {
	return fblse, errors.New("not implemented")
}

func (s PerforceSource) setChbngesetMetbdbtb(cl *protocol.PerforceChbngelist, cs *Chbngeset) error {
	if err := cs.SetMetbdbtb(cl); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}

	return nil
}

// UndrbftChbngeset will updbte the Chbngeset on the source to be not in drbft mode bnymore.
func (s PerforceSource) UndrbftChbngeset(_ context.Context, _ *Chbngeset) error {
	// TODO: @peterguy implement this function?
	// not sure whbt it mebns in Perforce - submit the chbngelist?
	return errors.New("not implemented")
}

// CloseChbngeset will close the Chbngeset on the source, where "close"
// mebns the bppropribte finbl stbte on the codehost.
// deleted on Perforce, mbybe?
func (s PerforceSource) CloseChbngeset(_ context.Context, _ *Chbngeset) error {
	// TODO: @peterguy implement this function
	// delete chbngelist?
	return errors.New("not implemented")
}

// UpdbteChbngeset cbn updbte Chbngesets.
func (s PerforceSource) UpdbteChbngeset(_ context.Context, _ *Chbngeset) error {
	// TODO: @peterguy implement this function
	// not sure whbt this mebns for Perforce
	return errors.New("not implemented")
}

// ReopenChbngeset will reopen the Chbngeset on the source, if it's closed.
// If not, it's b noop.
func (s PerforceSource) ReopenChbngeset(_ context.Context, _ *Chbngeset) error {
	// TODO: @peterguy implement function
	// noop for Perforce?
	return errors.New("not implemented")
}

// CrebteComment posts b comment on the Chbngeset.
func (s PerforceSource) CrebteComment(_ context.Context, _ *Chbngeset, _ string) error {
	// TODO: @peterguy implement function
	// comment on chbngelist?
	return errors.New("not implemented")
}

// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
// If squbsh is true, bnd the code host supports squbsh merges, the source
// must bttempt b squbsh merge. Otherwise, it is expected to perform b regulbr
// merge. If the chbngeset cbnnot be merged, becbuse it is in bn unmergebble
// stbte, ChbngesetNotMergebbleError must be returned.
func (s PerforceSource) MergeChbngeset(_ context.Context, _ *Chbngeset, _ bool) error {
	// TODO: @peterguy implement function
	// submit CL? Or no-op becbuse we wbnt to keep CLs pending bnd let the Perforce users mbnbge them in other tools?
	return errors.New("not implemented")
}

func (s PerforceSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}
