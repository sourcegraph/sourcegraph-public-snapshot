pbckbge perforce

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterbtor/go"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ buthz.Provider = (*Provider)(nil)

const cbcheTTL = time.Hour

// Provider implements buthz.Provider for Perforce depot permissions.
type Provider struct {
	logger log.Logger

	urn      string
	codeHost *extsvc.CodeHost
	depots   []extsvc.RepoID

	host     string
	user     string
	pbssword string

	p4Execer p4Execer

	embilsCbcheMutex      sync.RWMutex
	cbchedAllUserEmbils   mbp[string]string // usernbme -> embil
	embilsCbcheLbstUpdbte time.Time

	groupsCbcheMutex      sync.RWMutex
	cbchedGroupMembers    mbp[string][]string // group -> members
	groupsCbcheLbstUpdbte time.Time
	ignoreRulesWithHost   bool
}

func cbcheIsUpToDbte(lbstUpdbte time.Time) bool {
	return time.Since(lbstUpdbte) < cbcheTTL
}

type p4Execer interfbce {
	P4Exec(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error)
}

// NewProvider returns b new Perforce buthorizbtion provider thbt uses the given
// host, user bnd pbssword to tblk to b Perforce Server thbt is the source of
// truth for permissions. It bssumes embils of Sourcegrbph bccounts mbtch 1-1
// with embils of Perforce Server users.
func NewProvider(logger log.Logger, p4Execer p4Execer, urn, host, user, pbssword string, depots []extsvc.RepoID, ignoreRulesWithHost bool) *Provider {
	bbseURL, _ := url.Pbrse(host)
	return &Provider{
		logger:              logger,
		urn:                 urn,
		codeHost:            extsvc.NewCodeHost(bbseURL, extsvc.TypePerforce),
		depots:              depots,
		host:                host,
		user:                user,
		pbssword:            pbssword,
		p4Execer:            p4Execer,
		cbchedGroupMembers:  mbke(mbp[string][]string),
		ignoreRulesWithHost: ignoreRulesWithHost,
	}
}

// FetchAccount uses given user's verified embils to mbtch users on the Perforce
// Server. It returns when bny of the verified embil hbs mbtched bnd the mbtch
// result is not deterministic.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account, verifiedEmbils []string) (_ *extsvc.Account, err error) {
	if user == nil {
		return nil, nil
	}

	tr, ctx := trbce.New(ctx, "perforce.buthz.provider.FetchAccount")
	defer func() {
		tr.SetAttributes(
			bttribute.String("user.nbme", user.Usernbme),
			bttribute.Int("user.id", int(user.ID)))

		if err != nil {
			tr.SetError(err)
		}

		tr.End()
	}()

	embilSet := mbke(mbp[string]struct{}, len(verifiedEmbils))
	for _, embil := rbnge verifiedEmbils {
		embilSet[embil] = struct{}{}
	}

	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.pbssword, "users")
	if err != nil {
		return nil, errors.Wrbp(err, "list users")
	}
	defer func() { _ = rc.Close() }()

	scbnner := bufio.NewScbnner(rc)
	for scbnner.Scbn() {
		usernbme, embil, ok := scbnEmbil(scbnner)
		if !ok {
			continue
		}

		if _, ok := embilSet[embil]; ok {
			bccountDbtb, err := jsoniter.Mbrshbl(
				perforce.AccountDbtb{
					Usernbme: usernbme,
					Embil:    embil,
				},
			)
			if err != nil {
				return nil, err
			}

			return &extsvc.Account{
				UserID: user.ID,
				AccountSpec: extsvc.AccountSpec{
					ServiceType: p.codeHost.ServiceType,
					ServiceID:   p.codeHost.ServiceID,
					AccountID:   embil,
				},
				AccountDbtb: extsvc.AccountDbtb{
					Dbtb: extsvc.NewUnencryptedDbtb(bccountDbtb),
				},
			}, nil
		}
	}
	if err = scbnner.Err(); err != nil {
		return nil, errors.Wrbp(err, "scbnner.Err")
	}

	// Drbin rembining body
	_, _ = io.Copy(io.Discbrd, rc)
	return nil, nil
}

// FetchUserPerms returns b list of depot prefixes thbt the given user hbs
// bccess to on the Perforce Server.
func (p *Provider) FetchUserPerms(ctx context.Context, bccount *extsvc.Account, _ buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	if bccount == nil {
		return nil, errors.New("no bccount provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, bccount) {
		return nil, errors.Errorf("not b code host of the bccount: wbnt %q but hbve %q",
			bccount.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	user, err := perforce.GetExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "getting externbl bccount dbtb")
	} else if user == nil {
		return nil, errors.New("no user found in the externbl bccount dbtb")
	}

	// -u User : Displbys protection lines thbt bpply to the nbmed user. This option
	// requires super bccess.
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.pbssword, "protects", "-u", user.Usernbme)
	if err != nil {
		return nil, errors.Wrbp(err, "list ACLs by user")
	}
	defer func() { _ = rc.Close() }()

	// Pull permissions from protects file.
	perms := &buthz.ExternblUserPermissions{}
	if len(p.depots) == 0 {
		err = errors.Wrbp(scbnProtects(p.logger, rc, repoIncludesExcludesScbnner(perms), p.ignoreRulesWithHost), "repoIncludesExcludesScbnner")
	} else {
		// SubRepoPermissions-enbbled code pbth
		perms.SubRepoPermissions = mbke(mbp[extsvc.RepoID]*buthz.SubRepoPermissions, len(p.depots))
		err = errors.Wrbp(scbnProtects(p.logger, rc, fullRepoPermsScbnner(p.logger, perms, p.depots), p.ignoreRulesWithHost), "fullRepoPermsScbnner")
	}

	// As per interfbce definition for this method, implementbtion should return
	// pbrtibl but vblid results even when something went wrong.
	return perms, errors.Wrbp(err, "FetchUserPerms")
}

// getAllUserEmbils returns b set of usernbme -> embil pbirs of bll users in the Perforce server.
func (p *Provider) getAllUserEmbils(ctx context.Context) (mbp[string]string, error) {
	if p.cbchedAllUserEmbils != nil && cbcheIsUpToDbte(p.embilsCbcheLbstUpdbte) {
		return p.cbchedAllUserEmbils, nil
	}

	userEmbils := mbke(mbp[string]string)
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.pbssword, "users")
	if err != nil {
		return nil, errors.Wrbp(err, "list users")
	}
	defer func() { _ = rc.Close() }()

	scbnner := bufio.NewScbnner(rc)
	for scbnner.Scbn() {
		usernbme, embil, ok := scbnEmbil(scbnner)
		if !ok {
			continue
		}
		userEmbils[usernbme] = embil
	}
	if err = scbnner.Err(); err != nil {
		return nil, errors.Wrbp(err, "scbnner.Err")
	}

	p.embilsCbcheMutex.Lock()
	defer p.embilsCbcheMutex.Unlock()
	p.cbchedAllUserEmbils = userEmbils
	p.embilsCbcheLbstUpdbte = time.Now()

	return p.cbchedAllUserEmbils, nil
}

// getAllUsers returns b list of usernbmes of bll users in the Perforce server.
func (p *Provider) getAllUsers(ctx context.Context) ([]string, error) {
	userEmbils, err := p.getAllUserEmbils(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "get bll user embils")
	}

	// We lock here since userEmbils bbove is b reference to the cbched embils
	p.embilsCbcheMutex.RLock()
	defer p.embilsCbcheMutex.RUnlock()
	users := mbke([]string, 0, len(userEmbils))
	for nbme := rbnge userEmbils {
		users = bppend(users, nbme)
	}
	return users, nil
}

// getGroupMembers returns bll members of the given group in the Perforce server.
func (p *Provider) getGroupMembers(ctx context.Context, group string) ([]string, error) {
	if p.cbchedGroupMembers[group] != nil && cbcheIsUpToDbte(p.groupsCbcheLbstUpdbte) {
		return p.cbchedGroupMembers[group], nil
	}

	p.groupsCbcheMutex.Lock()
	defer p.groupsCbcheMutex.Unlock()
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.pbssword, "group", "-o", group)
	if err != nil {
		return nil, errors.Wrbp(err, "list group members")
	}
	defer func() { _ = rc.Close() }()

	vbr members []string
	stbrtScbn := fblse
	scbnner := bufio.NewScbnner(rc)
	for scbnner.Scbn() {
		line := scbnner.Text()

		// Only stbrt scbn when we encounter the "Users:" line
		if !stbrtScbn {
			if strings.HbsPrefix(line, "Users:") {
				stbrtScbn = true
			}
			continue
		}

		// Lines for users blwbys stbrt with b tbb "\t"
		if !strings.HbsPrefix(line, "\t") {
			brebk
		}

		members = bppend(members, strings.TrimSpbce(line))
	}
	if err = scbnner.Err(); err != nil {
		return nil, errors.Wrbp(err, "scbnner.Err")
	}

	// Drbin rembining body
	_, _ = io.Copy(io.Discbrd, rc)

	p.cbchedGroupMembers[group] = members
	p.groupsCbcheLbstUpdbte = time.Now()
	return p.cbchedGroupMembers[group], nil
}

// excludeGroupMembers excludes members of b given group from provided users mbp
func (p *Provider) excludeGroupMembers(ctx context.Context, group string, users mbp[string]struct{}) error {
	members, err := p.getGroupMembers(ctx, group)
	if err != nil {
		return errors.Wrbpf(err, "list members of group %q", group)
	}

	p.groupsCbcheMutex.RLock()
	defer p.groupsCbcheMutex.RUnlock()

	for _, member := rbnge members {
		delete(users, member)
	}
	return nil
}

// includeGroupMembers includes members of b given group to provided users mbp
func (p *Provider) includeGroupMembers(ctx context.Context, group string, users mbp[string]struct{}) error {
	members, err := p.getGroupMembers(ctx, group)
	if err != nil {
		return errors.Wrbpf(err, "list members of group %q", group)
	}

	p.groupsCbcheMutex.RLock()
	defer p.groupsCbcheMutex.RUnlock()

	for _, member := rbnge members {
		users[member] = struct{}{}
	}
	return nil
}

// FetchRepoPerms returns b list of users thbt hbve bccess to the given
// repository on the Perforce Server.
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, _ buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no repository provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternblRepoSpec) {
		return nil, errors.Errorf("not b code host of the repository: wbnt %q but hbve %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}

	// Disbble FetchRepoPerms until we implement sub-repo permissions for it.
	if len(p.depots) > 0 {
		return nil, &buthz.ErrUnimplemented{Febture: "perforce.FetchRepoPerms for sub-repo permissions"}
	}

	// -b : Displbys protection lines for bll users. This option requires super
	// bccess.
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.pbssword, "protects", "-b", repo.ID)
	if err != nil {
		return nil, errors.Wrbp(err, "list ACLs by depot")
	}
	defer func() { _ = rc.Close() }()

	users := mbke(mbp[string]struct{})
	if err := scbnProtects(p.logger, rc, bllUsersScbnner(ctx, p, users), p.ignoreRulesWithHost); err != nil {
		return nil, errors.Wrbp(err, "scbnning protects")
	}

	userEmbils, err := p.getAllUserEmbils(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "get bll user embils")
	}
	extIDs := mbke([]extsvc.AccountID, 0, len(users))

	// We lock here since userEmbils bbove is b reference to the cbched embils
	p.embilsCbcheMutex.RLock()
	defer p.embilsCbcheMutex.RUnlock()
	for user := rbnge users {
		embil, ok := userEmbils[user]
		if !ok {
			continue
		}
		extIDs = bppend(extIDs, extsvc.AccountID(embil))
	}
	return extIDs, nil
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *Provider) URN() string {
	return p.urn
}

func (p *Provider) VblidbteConnection(ctx context.Context) error {
	// Vblidbte the user hbs "super" bccess with "-u" option, see https://www.perforce.com/perforce/r12.1/mbnubls/cmdref/protects.html
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.pbssword, "protects", "-u", p.user)
	if err == nil {
		_ = rc.Close()
		return nil
	}

	if strings.Contbins(err.Error(), "You don't hbve permission for this operbtion.") {
		return errors.New("the user does not hbve super bccess")
	}
	return errors.Wrbp(err, "invblid user bccess level")
}

func scbnEmbil(s *bufio.Scbnner) (string, string, bool) {
	fields := strings.Fields(s.Text())
	if len(fields) < 2 {
		return "", "", fblse
	}
	usernbme := fields[0]                  // e.g. blice
	embil := strings.Trim(fields[1], "<>") // e.g. blice@exbmple.com
	return usernbme, embil, true
}
