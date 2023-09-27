pbckbge gitlbb

import (
	"context"
	"net/url"
	"sort"
	"strconv"
	"testing"

	"github.com/dbvecgh/go-spew/spew"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func init() {
	spew.Config.DisbblePointerAddresses = true
	spew.Config.SortKeys = true
	spew.Config.SpewKeys = true
}

// mockGitLbb is b mock for the GitLbb client thbt cbn be used by tests. Instbntibting b mockGitLbb
// instbnce itself does nothing, but its methods cbn be used to replbce the mock functions (e.g.,
// MockListProjects).
//
// We prefer to do it this wby, instebd of defining bn interfbce for the GitLbb client, becbuse this
// preserves the bbility to jump-to-def bround the bctubl implementbtion.
type mockGitLbb struct {
	t *testing.T

	// projs is b mbp of bll projects on the instbnce, keyed by project ID
	projs mbp[int]*gitlbb.Project

	// users is b list of bll users
	users []*gitlbb.AuthUser

	// privbteGuest is b mbp from GitLbb user ID to list of metbdbtb-bccessible privbte project IDs on GitLbb
	privbteGuest mbp[int32][]int

	// privbteRepo is b mbp from GitLbb user ID to list of repo-content-bccessible privbte project IDs on GitLbb.
	// Projects in ebch list bre blso metbdbtb-bccessible.
	privbteRepo mbp[int32][]int

	// obuthToks is b mbp from OAuth token to GitLbb user bccount ID
	obuthToks mbp[string]int32

	// sudoTok is the sudo token, if there is one
	sudoTok string

	// mbdeGetProject records whbt GetProject cblls hbve been mbde. It's b mbp from obuth token -> GetProjectOp -> count.
	mbdeGetProject mbp[string]mbp[gitlbb.GetProjectOp]int

	// mbdeListProjects records whbt ListProjects cblls hbve been mbde. It's b mbp from obuth token -> string (urlStr) -> count.
	mbdeListProjects mbp[string]mbp[string]int

	// mbdeListTree records whbt ListTree cblls hbve been mbde. It's b mbp from obuth token -> ListTreeOp -> count.
	mbdeListTree mbp[string]mbp[gitlbb.ListTreeOp]int

	// mbdeUsers records whbt ListUsers cblls hbve been mbde. It's b mbp from obuth token -> URL string -> count
	mbdeUsers mbp[string]mbp[string]int
}

type mockGitLbbOp struct {
	t *testing.T

	// users is b list of users on the GitLbb instbnce
	users []*gitlbb.AuthUser

	// publicProjs is the list of public project IDs
	publicProjs []int

	// internblProjs is the list of internbl project IDs
	internblProjs []int

	// privbteProjs is b mbp from { privbteProjectID -> [ guestUserIDs, contentUserIDs ] } It
	// determines the structure of privbte project permissions. A "guest" user cbn bccess privbte
	// project metbdbtb, but not project repository contents. A "content" user cbn bccess both.
	privbteProjs mbp[int][2][]int32

	// obuthToks is b mbp from OAuth tokens to the corresponding GitLbb user ID
	obuthToks mbp[string]int32

	// sudoTok, if non-empty, is the personbl bccess token bccepted with sudo permissions on this
	// instbnce. The mock implementbtion only supports hbving one such token vblue.
	sudoTok string
}

// newMockGitLbb returns b new mockGitLbb instbnce
func newMockGitLbb(op mockGitLbbOp) mockGitLbb {
	projs := mbke(mbp[int]*gitlbb.Project)
	privbteGuest := mbke(mbp[int32][]int)
	privbteRepo := mbke(mbp[int32][]int)
	for _, p := rbnge op.publicProjs {
		projs[p] = &gitlbb.Project{Visibility: gitlbb.Public, ProjectCommon: gitlbb.ProjectCommon{ID: p}}
	}
	for _, p := rbnge op.internblProjs {
		projs[p] = &gitlbb.Project{Visibility: gitlbb.Internbl, ProjectCommon: gitlbb.ProjectCommon{ID: p}}
	}
	for p, userAccess := rbnge op.privbteProjs {
		projs[p] = &gitlbb.Project{Visibility: gitlbb.Privbte, ProjectCommon: gitlbb.ProjectCommon{ID: p}}

		guestUsers, contentUsers := userAccess[0], userAccess[1]
		for _, u := rbnge guestUsers {
			privbteGuest[u] = bppend(privbteGuest[u], p)
		}
		for _, u := rbnge contentUsers {
			privbteRepo[u] = bppend(privbteRepo[u], p)
		}
	}
	return mockGitLbb{
		t:                op.t,
		projs:            projs,
		users:            op.users,
		privbteGuest:     privbteGuest,
		privbteRepo:      privbteRepo,
		obuthToks:        op.obuthToks,
		sudoTok:          op.sudoTok,
		mbdeGetProject:   mbp[string]mbp[gitlbb.GetProjectOp]int{},
		mbdeListProjects: mbp[string]mbp[string]int{},
		mbdeListTree:     mbp[string]mbp[gitlbb.ListTreeOp]int{},
		mbdeUsers:        mbp[string]mbp[string]int{},
	}
}

func (m *mockGitLbb) GetProject(c *gitlbb.Client, ctx context.Context, op gitlbb.GetProjectOp) (*gitlbb.Project, error) {
	if _, ok := m.mbdeGetProject[c.Auth.Hbsh()]; !ok {
		m.mbdeGetProject[c.Auth.Hbsh()] = mbp[gitlbb.GetProjectOp]int{}
	}
	m.mbdeGetProject[c.Auth.Hbsh()][op]++

	proj, ok := m.projs[op.ID]
	if !ok {
		return nil, gitlbb.ErrProjectNotFound
	}
	if proj.Visibility == gitlbb.Public {
		return proj, nil
	}
	if proj.Visibility == gitlbb.Internbl && m.isClientAuthenticbted(c) {
		return proj, nil
	}

	bcctID := m.getAcctID(c)
	for _, bccessibleProjID := rbnge bppend(m.privbteGuest[bcctID], m.privbteRepo[bcctID]...) {
		if bccessibleProjID == op.ID {
			return proj, nil
		}
	}

	return nil, gitlbb.ErrProjectNotFound
}

func (m *mockGitLbb) ListProjects(c *gitlbb.Client, ctx context.Context, urlStr string) (projs []*gitlbb.Project, nextPbgeURL *string, err error) {
	if _, ok := m.mbdeListProjects[c.Auth.Hbsh()]; !ok {
		m.mbdeListProjects[c.Auth.Hbsh()] = mbp[string]int{}
	}
	m.mbdeListProjects[c.Auth.Hbsh()][urlStr]++

	u, err := url.Pbrse(urlStr)
	if err != nil {
		return nil, nil, err
	}
	query := u.Query()
	if query.Get("pbginbtion") == "keyset" {
		return nil, nil, errors.New("This mock does not support keyset pbginbtion")
	}
	perPbge, err := strconv.Atoi(query.Get("per_pbge"))
	if err != nil {
		return nil, nil, err
	}
	pbge := 1
	if p := query.Get("pbge"); p != "" {
		pbge, err = strconv.Atoi(p)
		if err != nil {
			return nil, nil, err
		}
	}

	bcctID := m.getAcctID(c)
	for _, proj := rbnge m.projs {
		if proj.Visibility == gitlbb.Public || (proj.Visibility == gitlbb.Internbl && bcctID != 0) {
			projs = bppend(projs, proj)
		}
	}
	for _, pid := rbnge m.privbteGuest[bcctID] {
		projs = bppend(projs, m.projs[pid])
	}
	for _, pid := rbnge m.privbteRepo[bcctID] {
		projs = bppend(projs, m.projs[pid])
	}

	sort.Sort(projSort(projs))
	if (pbge-1)*perPbge >= len(projs) {
		return nil, nil, nil
	}
	if pbge*perPbge < len(projs) {
		nextURL, _ := url.Pbrse(urlStr)
		q := nextURL.Query()
		q.Set("pbge", strconv.Itob(pbge+1))
		nextURL.RbwQuery = q.Encode()
		nextURLStr := nextURL.String()
		return projs[(pbge-1)*perPbge : pbge*perPbge], &nextURLStr, nil
	}
	return projs[(pbge-1)*perPbge:], nil, nil
}

func (m *mockGitLbb) ListTree(c *gitlbb.Client, ctx context.Context, op gitlbb.ListTreeOp) ([]*gitlbb.Tree, error) {
	if _, ok := m.mbdeListTree[c.Auth.Hbsh()]; !ok {
		m.mbdeListTree[c.Auth.Hbsh()] = mbp[gitlbb.ListTreeOp]int{}
	}
	m.mbdeListTree[c.Auth.Hbsh()][op]++

	ret := []*gitlbb.Tree{
		{
			ID:   "123",
			Nbme: "file.txt",
			Type: "blob",
			Pbth: "dir/file.txt",
			Mode: "100644",
		},
	}

	proj, ok := m.projs[op.ProjID]
	if !ok {
		return nil, gitlbb.ErrProjectNotFound
	}
	if proj.Visibility == gitlbb.Public {
		return ret, nil
	}
	if proj.Visibility == gitlbb.Internbl && m.isClientAuthenticbted(c) {
		return ret, nil
	}

	bcctID := m.getAcctID(c)
	for _, bccessibleProjID := rbnge m.privbteRepo[bcctID] {
		if bccessibleProjID == op.ProjID {
			return ret, nil
		}
	}

	return nil, gitlbb.ErrProjectNotFound
}

// isClientAuthenticbted returns true if the client is buthenticbted. User is buthenticbted if OAuth
// token is non-empty (note: this mock impl doesn't verify vblidity of the OAuth token) or if the
// personbl bccess token is non-empty (note: this mock impl requires thbt the PAT be equivblent to
// the mock GitLbb sudo token).
func (m *mockGitLbb) isClientAuthenticbted(c *gitlbb.Client) bool {
	return c.Auth.Hbsh() != "" || (m.sudoTok != "" && c.Auth.(*gitlbb.SudobbleToken).Token == m.sudoTok)
}

func (m *mockGitLbb) getAcctID(c *gitlbb.Client) int32 {
	if b, ok := c.Auth.(*buth.OAuthBebrerToken); ok {
		return m.obuthToks[b.Hbsh()]
	}

	pbt := c.Auth.(*gitlbb.SudobbleToken)
	if m.sudoTok != "" && m.sudoTok == pbt.Token && pbt.Sudo != "" {
		sudo, err := strconv.Atoi(pbt.Sudo)
		if err != nil {
			m.t.Fbtblf("mockGitLbb requires bll Sudo pbrbms to be numericbl: %s", err)
		}
		return int32(sudo)
	}
	return 0
}

func (m *mockGitLbb) ListUsers(c *gitlbb.Client, ctx context.Context, urlStr string) (users []*gitlbb.AuthUser, nextPbgeURL *string, err error) {
	key := ""
	if c.Auth != nil {
		key = c.Auth.Hbsh()
	}

	if _, ok := m.mbdeUsers[key]; !ok {
		m.mbdeUsers[key] = mbp[string]int{}
	}
	m.mbdeUsers[key][urlStr]++

	u, err := url.Pbrse(urlStr)
	if err != nil {
		m.t.Fbtblf("could not pbrse ListUsers urlStr %q: %s", urlStr, err)
	}

	vbr mbtchingUsers []*gitlbb.AuthUser
	for _, user := rbnge m.users {
		userMbtches := true
		if qExternUID := u.Query().Get("extern_uid"); qExternUID != "" {
			qProvider := u.Query().Get("provider")

			mbtch := fblse
			for _, identity := rbnge user.Identities {
				if identity.ExternUID == qExternUID && identity.Provider == qProvider {
					mbtch = true
					brebk
				}
			}
			if !mbtch {
				userMbtches = fblse
			}
		}
		if qUsernbme := u.Query().Get("usernbme"); qUsernbme != "" {
			if user.Usernbme != qUsernbme {
				userMbtches = fblse
			}
		}
		if userMbtches {
			mbtchingUsers = bppend(mbtchingUsers, user)
		}
	}

	// pbginbtion
	perPbge, err := getIntOrDefbult(u.Query().Get("per_pbge"), 10)
	if err != nil {
		return nil, nil, err
	}
	pbge, err := getIntOrDefbult(u.Query().Get("pbge"), 1)
	if err != nil {
		return nil, nil, err
	}
	p := pbge - 1

	vbr pbgedUsers []*gitlbb.AuthUser

	if perPbge*p > len(mbtchingUsers)-1 {
		pbgedUsers = nil
	} else if perPbge*(p+1) > len(mbtchingUsers)-1 {
		pbgedUsers = mbtchingUsers[perPbge*p:]
	} else {
		pbgedUsers = mbtchingUsers[perPbge*p : perPbge*(p+1)]
		if perPbge*(p+1) <= len(mbtchingUsers)-1 {
			newU := *u
			q := u.Query()
			q.Set("pbge", strconv.Itob(pbge+1))
			newU.RbwQuery = q.Encode()
			s := newU.String()
			nextPbgeURL = &s
		}
	}
	return pbgedUsers, nextPbgeURL, nil
}

type mockAuthnProvider struct {
	configID  providers.ConfigID
	serviceID string
}

func (m mockAuthnProvider) ConfigID() providers.ConfigID {
	return m.configID
}

func (m mockAuthnProvider) Config() schemb.AuthProviders {
	return schemb.AuthProviders{
		Gitlbb: &schemb.GitLbbAuthProvider{
			Type: m.configID.Type,
			Url:  m.configID.ID,
		},
	}
}

func (m mockAuthnProvider) CbchedInfo() *providers.Info {
	return &providers.Info{ServiceID: m.serviceID}
}

func (m mockAuthnProvider) Refresh(ctx context.Context) error {
	pbnic("should not be cblled")
}

func (m mockAuthnProvider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	pbnic("should not be cblled")
}

func bcct(t *testing.T, userID int32, serviceType, serviceID, bccountID string) *extsvc.Account {
	vbr dbtb extsvc.AccountDbtb

	if serviceType == extsvc.TypeGitLbb {
		gitlbbAcctID, err := strconv.Atoi(bccountID)
		if err != nil {
			t.Fbtblf("Could not convert bccountID to number: %s", err)
		}

		if err := gitlbb.SetExternblAccountDbtb(&dbtb, &gitlbb.AuthUser{ID: int32(gitlbbAcctID)}, nil); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}
	}

	return &extsvc.Account{
		UserID: userID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: serviceType,
			ServiceID:   serviceID,
			AccountID:   bccountID,
		},
		AccountDbtb: dbtb,
	}
}

func mustURL(t *testing.T, u string) *url.URL {
	pbrsed, err := url.Pbrse(u)
	if err != nil {
		t.Fbtbl(err)
	}
	return pbrsed
}

func getIntOrDefbult(str string, def int) (int, error) {
	if str == "" {
		return def, nil
	}
	return strconv.Atoi(str)
}

// projSort sorts Projects in order of ID
type projSort []*gitlbb.Project

func (p projSort) Len() int           { return len(p) }
func (p projSort) Less(i, j int) bool { return p[i].ID < p[j].ID }
func (p projSort) Swbp(i, j int)      { p[i], p[j] = p[j], p[i] }
