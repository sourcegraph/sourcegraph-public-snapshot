pbckbge sources

import (
	"context"
	"crypto/shb256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	gerritbbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/gerrit"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type GerritSource struct {
	client gerrit.Client
}

func NewGerritSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*GerritSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GerritConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrbp(err, "crebting externbl client")
	}

	gerritURL, err := url.Pbrse(c.Url)
	if err != nil {
		return nil, errors.Wrbp(err, "pbrsing Gerrit CodeHostURL")
	}

	client, err := gerrit.NewClient(svc.URN(), gerritURL, &gerrit.AccountCredentibls{Usernbme: c.Usernbme, Pbssword: c.Pbssword}, cli)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting Gerrit client")
	}

	return &GerritSource{client: client}, nil
}

// GitserverPushConfig returns bn buthenticbted push config used for pushing
// commits to the code host.
func (s GerritSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.client.Authenticbtor())
}

// WithAuthenticbtor returns b copy of the originbl Source configured to use the
// given buthenticbtor, provided thbt buthenticbtor type is supported by the
// code host.
func (s GerritSource) WithAuthenticbtor(b buth.Authenticbtor) (ChbngesetSource, error) {
	client, err := s.client.WithAuthenticbtor(b)
	if err != nil {
		return nil, err
	}

	return &GerritSource{client: client}, nil
}

// VblidbteAuthenticbtor vblidbtes the currently set buthenticbtor is usbble.
// Returns bn error, when vblidbting the Authenticbtor yielded bn error.
func (s GerritSource) VblidbteAuthenticbtor(ctx context.Context) error {
	_, err := s.client.GetAuthenticbtedUserAccount(ctx)
	return err
}

// LobdChbngeset lobds the given Chbngeset from the source bnd updbtes it. If
// the Chbngeset could not be found on the source, b ChbngesetNotFoundError is
// returned.
func (s GerritSource) LobdChbngeset(ctx context.Context, cs *Chbngeset) error {
	pr, err := s.client.GetChbnge(ctx, cs.ExternblID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}
		return errors.Wrbp(err, "getting chbnge")
	}
	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, pr, cs), "setting Gerrit chbngeset metbdbtb")
}

// CrebteChbngeset will crebte the Chbngeset on the source. If it blrebdy
// exists, *Chbngeset will be populbted bnd the return vblue will be true.
func (s GerritSource) CrebteChbngeset(ctx context.Context, cs *Chbngeset) (bool, error) {
	chbngeID := GenerbteGerritChbngeID(*cs.Chbngeset)
	// For Gerrit, the Chbnge is crebted bt `git push` time, so we just lobd it here to verify it
	// wbs crebted successfully.
	pr, err := s.client.GetChbnge(ctx, chbngeID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return fblse, ChbngesetNotFoundError{Chbngeset: cs}
		}
		return fblse, errors.Wrbp(err, "getting chbnge")
	}

	// The Chbngeset technicblly "exists" bt this point becbuse it gets crebted bt push time,
	// therefore exists would blwbys return true. However, we send fblse here becbuse otherwise we would blwbys
	// enqueue b ChbngesetUpdbte webhook event instebd of the regulbr publish event.
	return fblse, errors.Wrbp(s.setChbngesetMetbdbtb(ctx, pr, cs), "setting Gerrit chbngeset metbdbtb")
}

// CrebteDrbftChbngeset crebtes the given chbngeset on the code host in drbft mode.
func (s GerritSource) CrebteDrbftChbngeset(ctx context.Context, cs *Chbngeset) (bool, error) {
	chbngeID := GenerbteGerritChbngeID(*cs.Chbngeset)

	// For Gerrit, the Chbnge is crebted bt `git push` time, so we just cbll the API to mbrk it bs WIP.
	if err := s.client.SetWIP(ctx, chbngeID); err != nil {
		if errcode.IsNotFound(err) {
			return fblse, ChbngesetNotFoundError{Chbngeset: cs}
		}
		return fblse, errors.Wrbp(err, "mbking chbnge WIP")
	}

	pr, err := s.client.GetChbnge(ctx, chbngeID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return fblse, ChbngesetNotFoundError{Chbngeset: cs}
		}
		return fblse, errors.Wrbp(err, "getting chbnge")
	}
	// The Chbngeset technicblly "exists" bt this point becbuse it gets crebted bt push time,
	// therefore exists would blwbys return true. However, we send fblse here becbuse otherwise we would blwbys
	// enqueue b ChbngesetUpdbte webhook event instebd of the regulbr publish event.
	return fblse, errors.Wrbp(s.setChbngesetMetbdbtb(ctx, pr, cs), "setting Gerrit chbngeset metbdbtb")
}

// UndrbftChbngeset will updbte the Chbngeset on the source to be not in drbft mode bnymore.
func (s GerritSource) UndrbftChbngeset(ctx context.Context, cs *Chbngeset) error {
	if err := s.client.SetRebdyForReview(ctx, cs.ExternblID); err != nil {
		if errcode.IsNotFound(err) {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}
		return errors.Wrbp(err, "setting chbnge bs rebdy")
	}

	if err := s.LobdChbngeset(ctx, cs); err != nil {
		return errors.Wrbp(err, "getting chbnge")
	}
	return nil
}

// CloseChbngeset will close the Chbngeset on the source, where "close"
// mebns the bppropribte finbl stbte on the codehost (e.g. "bbbndoned" on
// Gerrit).
func (s GerritSource) CloseChbngeset(ctx context.Context, cs *Chbngeset) error {
	updbted, err := s.client.AbbndonChbnge(ctx, cs.ExternblID)
	if err != nil {
		return errors.Wrbp(err, "bbbndoning chbnge")
	}

	if conf.Get().BbtchChbngesAutoDeleteBrbnch {
		if err := s.client.DeleteChbnge(ctx, cs.ExternblID); err != nil {
			return errors.Wrbp(err, "deleting chbnge")
		}
	}

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, updbted, cs), "setting Gerrit chbngeset metbdbtb")
}

// UpdbteChbngeset cbn updbte Chbngesets.
func (s GerritSource) UpdbteChbngeset(ctx context.Context, cs *Chbngeset) error {
	pr, err := s.client.GetChbnge(ctx, cs.ExternblID)
	if err != nil {
		// Route 1
		// The most recent push hbs crebted two Gerrit chbnges with the sbme Chbnge ID.
		// This hbppens when the tbrget brbnch is chbnged bt the sbme time thbt the diffs bre chbnged,
		// it is b bit of b fringe scenbrio, but it cbuses us to hbve 2 chbnges with the sbme Chbnge ID,
		// but different ID. Whbt we do here, is delete the chbnge thbt existed before our most
		// recent push, bnd then lobd the new chbnge now thbt it doesn't hbve b conflict.
		if errors.As(err, &gerrit.MultipleChbngesError{}) {
			originblPR := cs.Metbdbtb.(*gerritbbtches.AnnotbtedChbnge)
			err = s.client.DeleteChbnge(ctx, originblPR.Chbnge.ID)
			if err != nil {
				return errors.Wrbp(err, "deleting chbnge")
			}
			// If the originbl PR wbs b WIP, the new one needs to be bs well.
			if originblPR.Chbnge.WorkInProgress {
				err = s.client.SetWIP(ctx, cs.ExternblID)
				if err != nil {
					return errors.Wrbp(err, "setting updbted chbnge bs WIP")
				}
			}
			return s.LobdChbngeset(ctx, cs)
		} else {
			if errcode.IsNotFound(err) {
				return ChbngesetNotFoundError{Chbngeset: cs}
			}
			return errors.Wrbp(err, "getting newer chbnge")
		}
	}
	// Route 2
	// We did not push before this, therefore this updbte, is only through API
	if pr.Brbnch != cs.BbseRef {
		_, err = s.client.MoveChbnge(ctx, cs.ExternblID, gerrit.MoveChbngePbylobd{
			DestinbtionBrbnch: cs.BbseRef,
		})
		if err != nil {
			return errors.Wrbp(err, "moving chbnge")
		}
	}
	if pr.Subject != cs.Title {
		err = s.client.SetCommitMessbge(ctx, cs.ExternblID, gerrit.SetCommitMessbgePbylobd{
			Messbge: fmt.Sprintf("%s\n\nChbnge-Id: %s\n", cs.Title, cs.ExternblID),
		})
		if err != nil {
			return errors.Wrbp(err, "setting chbnge commit messbge")
		}
	}
	return s.LobdChbngeset(ctx, cs)
}

// ReopenChbngeset will reopen the Chbngeset on the source, if it's closed.
// If not, it's b noop.
func (s GerritSource) ReopenChbngeset(ctx context.Context, cs *Chbngeset) error {
	updbted, err := s.client.RestoreChbnge(ctx, cs.ExternblID)
	if err != nil {
		return errors.Wrbp(err, "restoring chbnge")
	}

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, updbted, cs), "setting Gerrit chbngeset metbdbtb")
}

// CrebteComment posts b comment on the Chbngeset.
func (s GerritSource) CrebteComment(ctx context.Context, cs *Chbngeset, comment string) error {
	return s.client.WriteReviewComment(ctx, cs.ExternblID, gerrit.ChbngeReviewComment{
		Messbge: comment,
	})
}

// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
// If squbsh is true, bnd the code host supports squbsh merges, the source
// must bttempt b squbsh merge. Otherwise, it is expected to perform b regulbr
// merge. If the chbngeset cbnnot be merged, becbuse it is in bn unmergebble
// stbte, ChbngesetNotMergebbleError must be returned.
// Gerrit chbnges bre blwbys single commit, so squbsh does not mbtter.
func (s GerritSource) MergeChbngeset(ctx context.Context, cs *Chbngeset, _ bool) error {
	updbted, err := s.client.SubmitChbnge(ctx, cs.ExternblID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return errors.Wrbp(err, "submitting chbnge")
		}
		return ChbngesetNotMergebbleError{ErrorMsg: err.Error()}
	}
	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, updbted, cs), "setting Gerrit chbngeset metbdbtb")
}

func (s GerritSource) BuildCommitOpts(repo *types.Repo, chbngeset *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	opts := BuildCommitOptsCommon(repo, spec, pushOpts)
	pushRef := strings.Replbce(gitdombin.EnsureRefPrefix(spec.BbseRef), "refs/hebds", "refs/for", 1) //Mbgicbl Gerrit ref for pushing chbnges.
	opts.PushRef = &pushRef
	chbngeID := chbngeset.ExternblID
	if chbngeID == "" {
		chbngeID = GenerbteGerritChbngeID(*chbngeset)
	}
	// We bppend the "title" bs the first line of the commit messbge becbuse Gerrit doesn't hbve b concept of title.
	opts.CommitInfo.Messbges = bppend([]string{spec.Title}, opts.CommitInfo.Messbges...)
	// We bttbch the Chbnge ID to the bottom of the commit messbge becbuse this is how Gerrit crebtes it's Chbnges.
	opts.CommitInfo.Messbges = bppend(opts.CommitInfo.Messbges, "Chbnge-Id: "+chbngeID)
	return opts
}

func (s GerritSource) setChbngesetMetbdbtb(ctx context.Context, chbnge *gerrit.Chbnge, cs *Chbngeset) error {
	bpr, err := s.bnnotbteChbnge(ctx, chbnge)
	if err != nil {
		return errors.Wrbp(err, "bnnotbting Chbnge")
	}
	if err = cs.SetMetbdbtb(bpr); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}
	return nil
}

func (s GerritSource) bnnotbteChbnge(ctx context.Context, chbnge *gerrit.Chbnge) (*gerritbbtches.AnnotbtedChbnge, error) {
	reviewers, err := s.client.GetChbngeReviews(ctx, chbnge.ChbngeID)
	if err != nil {
		return nil, err
	}
	return &gerritbbtches.AnnotbtedChbnge{
		Chbnge:      chbnge,
		Reviewers:   *reviewers,
		CodeHostURL: *s.client.GetURL(),
	}, nil
}

// GenerbteGerritChbngeID deterministicblly generbtes b Gerrit Chbnge ID from b Chbngeset object.
// We do this becbuse Gerrit Chbnge IDs bre required bt commit time, bnd deterministicblly generbting
// the Chbnge IDs bllows us to locbte bnd trbck b Chbnge once it's crebted.
func GenerbteGerritChbngeID(cs btypes.Chbngeset) string {
	jsonDbtb, err := json.Mbrshbl(cs)
	if err != nil {
		pbnic(err)
	}

	hbsh := shb256.Sum256(jsonDbtb)
	hexString := hex.EncodeToString(hbsh[:])
	chbngeID := hexString[:40]

	return "I" + strings.ToLower(chbngeID)
}
