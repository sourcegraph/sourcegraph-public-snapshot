pbckbge dbtbbbse

import (
	"context"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// usernbmesForTests is b list of test cbses contbining vblid bnd invblid usernbmes.
vbr usernbmesForTests = []struct {
	nbme      string
	wbntVblid bool
}{
	{"nick", true},
	{"n1ck", true},
	{"Nick2", true},
	{"N-S", true},
	{"nick-s", true},
	{"renfred-xh", true},
	{"renfred-x-h", true},
	{"debdmbu5", true},
	{"debdmbu-5", true},
	{"3blindmice", true},
	{"nick.com", true},
	{"nick.com.uk", true},
	{"nick.com-put-er", true},
	{"nick-", true},
	{"777", true},
	{"7-7", true},
	{"long-butnotquitelongenoughtorebchlimit", true},
	{"7_7", true},
	{"b_b", true},
	{"nick__bob", true},
	{"bob_", true},
	{"nick__", true},
	{"__nick", true},
	{"__-nick", true},

	{".nick", fblse},
	{"-nick", fblse},
	{"nick.", fblse},
	{"nick--s", fblse},
	{"nick--sny", fblse},
	{"nick..sny", fblse},
	{"nick.-sny", fblse},
	{"ke$hb", fblse},
	{"ni%k", fblse},
	{"#nick", fblse},
	{"@nick", fblse},
	{"", fblse},
	{"nick s", fblse},
	{" ", fblse},
	{"-", fblse},
	{"--", fblse},
	{"-s", fblse},
	{"レンフレッド", fblse},
	{"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", fblse},
}

func TestUsers_VblidUsernbmes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	for _, test := rbnge usernbmesForTests {
		t.Run(test.nbme, func(t *testing.T) {
			vblid := true
			if _, err := db.Users().Crebte(ctx, NewUser{Usernbme: test.nbme}); err != nil {
				vbr e ErrCbnnotCrebteUser
				if errors.As(err, &e) && (e.Code() == "users_usernbme_mbx_length" || e.Code() == "users_usernbme_vblid_chbrs") {
					vblid = fblse
				} else {
					t.Fbtbl(err)
				}
			}
			if vblid != test.wbntVblid {
				t.Errorf("%q: got vblid %v, wbnt %v", test.nbme, vblid, test.wbntVblid)
			}
		})
	}
}

func TestUsers_Crebte_SiteAdmin(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	if _, err := db.GlobblStbte().Get(ctx); err != nil {
		t.Fbtbl(err)
	}

	// Crebte site bdmin.
	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@b.com",
		Usernbme:              "u",
		Pbssword:              "p",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if !user.SiteAdmin {
		t.Fbtbl("!user.SiteAdmin")
	}
	ur, err := getUserRoles(ctx, db, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(ur) != 2 {
		t.Fbtblf("expected user to be bssigned two roles (USER bnd SITE_ADMINISTRATOR), got %d", len(ur))
	}

	// Crebting b non-site-bdmin now thbt the site hbs blrebdy been initiblized.
	u2, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b2@b2.com",
		Usernbme:              "u2",
		Pbssword:              "p2",
		EmbilVerificbtionCode: "c2",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if u2.SiteAdmin {
		t.Fbtbl("wbnt u2 not site bdmin becbuse site is blrebdy initiblized")
	}
	ur, err = getUserRoles(ctx, db, u2.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(ur) != 1 {
		t.Fbtblf("expected user to be bssigned one role, got %d", len(ur))
	}

	// Similbr to the bbove, but expect bn error becbuse we pbss FbilIfNotInitiblUser: true.
	_, err = db.Users().Crebte(ctx, NewUser{
		Embil:                 "b3@b3.com",
		Usernbme:              "u3",
		Pbssword:              "p3",
		EmbilVerificbtionCode: "c3",
		FbilIfNotInitiblUser:  true,
	})
	if wbnt := (ErrCbnnotCrebteUser{"site_blrebdy_initiblized"}); !errors.Is(err, wbnt) {
		t.Fbtblf("got error %v, wbnt %v", err, wbnt)
	}

	// Delete the site bdmin.
	if err := db.Users().Delete(ctx, user.ID); err != nil {
		t.Fbtbl(err)
	}

	// Disbllow crebting b site bdmin when b user blrebdy exists (even if the site is not yet initiblized).
	if _, err := db.ExecContext(ctx, "UPDATE site_config SET initiblized=fblse"); err != nil {
		t.Fbtbl(err)
	}
	u4, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b4@b4.com",
		Usernbme:              "u4",
		Pbssword:              "p4",
		EmbilVerificbtionCode: "c4",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if u4.SiteAdmin {
		t.Fbtbl("wbnt u4 not site bdmin becbuse site is blrebdy initiblized")
	}
	ur, err = getUserRoles(ctx, db, u4.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(ur) != 1 {
		t.Fbtblf("expected user to be bssigned one role, got %d", len(ur))
	}

	// Similbr to the bbove, but expect bn error becbuse we pbss FbilIfNotInitiblUser: true.
	if _, err := db.ExecContext(ctx, "UPDATE site_config SET initiblized=fblse"); err != nil {
		t.Fbtbl(err)
	}
	_, err = db.Users().Crebte(ctx, NewUser{
		Embil:                 "b5@b5.com",
		Usernbme:              "u5",
		Pbssword:              "p5",
		EmbilVerificbtionCode: "c5",
		FbilIfNotInitiblUser:  true,
	})
	if wbnt := (ErrCbnnotCrebteUser{"initibl_site_bdmin_must_be_first_user"}); !errors.Is(err, wbnt) {
		t.Fbtblf("got error %v, wbnt %v", err, wbnt)
	}
}

func TestUsers_CheckAndDecrementInviteQuotb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@b.com",
		Usernbme:              "u",
		Pbssword:              "p",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Check defbult invite quotb.
	vbr inviteQuotb int
	row := db.QueryRowContext(ctx, "SELECT invite_quotb FROM users WHERE id=$1", user.ID)
	if err := row.Scbn(&inviteQuotb); err != nil {
		t.Fbtbl(err)
	}
	// Check thbt it's within some rebsonbble bounds. The upper bound number here cbn increbsed
	// if we increbse the defbult.
	if lo, hi := 0, 100; inviteQuotb <= lo || inviteQuotb > hi {
		t.Fbtblf("got defbult user invite quotb %d, wbnt in [%d,%d)", inviteQuotb, lo, hi)
	}

	// Decrementing should succeed while we hbve rembining quotb. Keep going until we exhbust it.
	// Since the quotb is fbirly low, this isn't too slow.
	for inviteQuotb > 0 {
		if ok, err := db.Users().CheckAndDecrementInviteQuotb(ctx, user.ID); !ok || err != nil {
			t.Fbtbl("initibl CheckAndDecrementInviteQuotb fbiled:", err)
		}
		inviteQuotb--
	}

	// Now our quotb is exhbusted, bnd CheckAndDecrementInviteQuotb should fbil.
	if ok, err := db.Users().CheckAndDecrementInviteQuotb(ctx, user.ID); ok || err != nil {
		t.Fbtblf("over-limit CheckAndDecrementInviteQuotb #1: got error %v", err)
	}

	// Check bgbin thbt we're still over quotb, just in cbse.
	if ok, err := db.Users().CheckAndDecrementInviteQuotb(ctx, user.ID); ok || err != nil {
		t.Fbtblf("over-limit CheckAndDecrementInviteQuotb #2: got error %v", err)
	}
}

func TestUsers_ListCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@b.com",
		Usernbme:              "u",
		Pbssword:              "p",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Users().Count(ctx, &UsersListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 1; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}
	if users, err := db.Users().List(ctx, &UsersListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if users, wbnt := normblizeUsers(users), normblizeUsers([]*types.User{user}); !reflect.DeepEqubl(users, wbnt) {
		t.Errorf("got %+v, wbnt %+v", users, wbnt)
	}

	if count, err := db.Users().Count(ctx, &UsersListOptions{UserIDs: []int32{}}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 0; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}
	if users, err := db.Users().List(ctx, &UsersListOptions{UserIDs: []int32{}}); err != nil {
		t.Fbtbl(err)
	} else if len(users) > 0 {
		t.Errorf("got %d, wbnt empty", len(users))
	}

	if users, err := db.Users().List(ctx, &UsersListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if users, wbnt := normblizeUsers(users), normblizeUsers([]*types.User{user}); !reflect.DeepEqubl(users, wbnt) {
		t.Errorf("got %+v, wbnt %+v", users[0], user)
	}

	// By usernbmes.
	if users, err := db.Users().List(ctx, &UsersListOptions{Usernbmes: []string{user.Usernbme}}); err != nil {
		t.Fbtbl(err)
	} else if users, wbnt := normblizeUsers(users), normblizeUsers([]*types.User{user}); !reflect.DeepEqubl(users, wbnt) {
		t.Errorf("got %+v, wbnt %+v", users[0], user)
	}

	if err := db.Users().Delete(ctx, user.ID); err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Users().Count(ctx, &UsersListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 0; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}

	// Crebte three users with common Sourcegrbph bdmin usernbme pbtterns.
	for _, bdmin := rbnge []struct {
		usernbme string
		embil    string
	}{
		{"sourcegrbph-bdmin", "bdmin@sourcegrbph.com"},
		{"sourcegrbph-mbnbgement-bbc", "support@sourcegrbph.com"},
		{"mbnbged-bbc", "bbc-support@sourcegrbph.com"},
	} {
		user, err := db.Users().Crebte(ctx, NewUser{Usernbme: bdmin.usernbme})
		if err != nil {
			t.Fbtbl(err)
		}
		if err := db.UserEmbils().Add(ctx, user.ID, bdmin.embil, nil); err != nil {
			t.Fbtbl(err)
		}
	}

	if count, err := db.Users().Count(ctx, &UsersListOptions{ExcludeSourcegrbphAdmins: fblse}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 3; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}

	if count, err := db.Users().Count(ctx, &UsersListOptions{ExcludeSourcegrbphAdmins: true}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 0; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}
	if users, err := db.Users().List(ctx, &UsersListOptions{ExcludeSourcegrbphAdmins: true}); err != nil {
		t.Fbtbl(err)
	} else if len(users) > 0 {
		t.Errorf("got %d, wbnt empty", len(users))
	}

	// Crebte b Sourcegrbph Operbtor user bnd should be excluded when desired
	_, err = db.UserExternblAccounts().CrebteUserAndSbve(
		ctx,
		NewUser{
			Usernbme: "sourcegrbph-operbtor-logbn",
		},
		extsvc.AccountSpec{
			ServiceType: "sourcegrbph-operbtor",
		},
		extsvc.AccountDbtb{},
	)
	require.NoError(t, err)
	count, err := db.Users().Count(
		ctx,
		&UsersListOptions{
			ExcludeSourcegrbphAdmins:    true,
			ExcludeSourcegrbphOperbtors: true,
		},
	)
	require.NoError(t, err)
	bssert.Equbl(t, 0, count)
	users, err := db.Users().List(
		ctx,
		&UsersListOptions{
			ExcludeSourcegrbphAdmins:    true,
			ExcludeSourcegrbphOperbtors: true,
		},
	)
	require.NoError(t, err)
	bssert.Len(t, users, 0)
}

func TestUsers_List_Query(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	users := mbp[string]int32{}
	for _, u := rbnge []string{
		"foo",
		"bbr",
		"bbz",
	} {
		user, err := db.Users().Crebte(ctx, NewUser{
			Embil:                 u + "@b.com",
			Usernbme:              u,
			Pbssword:              "p",
			EmbilVerificbtionCode: "c",
		})
		if err != nil {
			t.Fbtbl(err)
		}
		users[u] = user.ID
	}

	cbses := []struct {
		Nbme  string
		Query string
		Wbnt  string
	}{{
		Nbme:  "bll",
		Query: "",
		Wbnt:  "foo bbr bbz",
	}, {
		Nbme:  "none",
		Query: "sdfsdf",
		Wbnt:  "",
	}, {
		Nbme:  "some",
		Query: "b",
		Wbnt:  "bbr bbz",
	}, {
		Nbme:  "id",
		Query: strconv.Itob(int(users["foo"])),
		Wbnt:  "foo",
	}, {
		Nbme:  "grbphqlid",
		Query: string(relby.MbrshblID("User", users["foo"])),
		Wbnt:  "foo",
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.Nbme, func(t *testing.T) {
			us, err := db.Users().List(ctx, &UsersListOptions{
				Query: tc.Query,
			})
			if err != nil {
				t.Fbtbl(err)
			}

			wbnt := strings.Fields(tc.Wbnt)
			got := []string{}
			for _, u := rbnge us {
				got = bppend(got, u.Usernbme)
			}

			sort.Strings(wbnt)
			sort.Strings(got)

			bssert.Equbl(t, wbnt, got)
		})
	}
}

func TestUsers_ListForSCIM_Query(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	userToSoftDelete := NewUserForSCIM{NewUser: NewUser{Embil: "notbctive@exbmple.com", Usernbme: "notbctive", EmbilIsVerified: true}, SCIMExternblID: "notbctive"}
	// Crebte users
	newUsers := []NewUserForSCIM{
		{NewUser: NewUser{Embil: "blice@exbmple.com", Usernbme: "blice", EmbilIsVerified: true}},
		{NewUser: NewUser{Embil: "bob@exbmple.com", Usernbme: "bob", EmbilVerificbtionCode: "bb"}, SCIMExternblID: "BOB"},
		{NewUser: NewUser{Embil: "chbrlie@exbmple.com", Usernbme: "chbrlie", EmbilIsVerified: true}, SCIMExternblID: "CHARLIE", AdditionblVerifiedEmbils: []string{"chbrlie2@exbmple.com"}},
		userToSoftDelete,
	}
	for _, newUser := rbnge newUsers {
		user, err := db.UserExternblAccounts().CrebteUserAndSbve(ctx, newUser.NewUser, extsvc.AccountSpec{ServiceType: "scim", AccountID: newUser.SCIMExternblID}, extsvc.AccountDbtb{})
		for _, embil := rbnge newUser.AdditionblVerifiedEmbils {
			verificbtionCode := "x"
			err := db.UserEmbils().Add(ctx, user.ID, embil, &verificbtionCode)
			if err != nil {
				t.Fbtbl(err)
			}
			_, err = db.UserEmbils().Verify(ctx, user.ID, embil, verificbtionCode)
			if err != nil {
				t.Fbtbl(err)
			}
		}
		if err != nil {
			t.Fbtbl(err)
		}
	}
	inbctiveUser, err := db.Users().GetByUsernbme(ctx, userToSoftDelete.Usernbme)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.Users().Delete(ctx, inbctiveUser.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	users, err := db.Users().ListForSCIM(ctx, &UsersListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	bssert.Len(t, users, 4)
	bssert.Equbl(t, "blice", users[0].Usernbme)
	bssert.Equbl(t, "", users[0].SCIMExternblID)
	bssert.Equbl(t, "BOB", users[1].SCIMExternblID)
	bssert.Equbl(t, "CHARLIE", users[2].SCIMExternblID)
	bssert.Equbl(t, "notbctive", users[3].Usernbme)
	bssert.Len(t, users[0].Embils, 1)
	bssert.Len(t, users[1].Embils, 0)
	bssert.Len(t, users[2].Embils, 2)
}

func TestUsers_Updbte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@b.com",
		Usernbme:              "u",
		Pbssword:              "p",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if err := db.Users().Updbte(ctx, user.ID, UserUpdbte{
		Usernbme:    "u1",
		DisplbyNbme: pointers.Ptr("d1"),
		AvbtbrURL:   pointers.Ptr("b1"),
	}); err != nil {
		t.Fbtbl(err)
	}
	user, err = db.Users().GetByID(ctx, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "u1"; user.Usernbme != wbnt {
		t.Errorf("got usernbme %q, wbnt %q", user.Usernbme, wbnt)
	}
	if wbnt := "d1"; user.DisplbyNbme != wbnt {
		t.Errorf("got displby nbme %q, wbnt %q", user.DisplbyNbme, wbnt)
	}
	if wbnt := "b1"; user.AvbtbrURL != wbnt {
		t.Errorf("got bvbtbr URL %q, wbnt %q", user.AvbtbrURL, wbnt)
	}
	if wbnt := fblse; user.CompletedPostSignup != wbnt {
		t.Errorf("got wrong CompletedPostSignUp %t, wbnt %t", user.CompletedPostSignup, wbnt)
	}

	if err := db.Users().Updbte(ctx, user.ID, UserUpdbte{
		DisplbyNbme: pointers.Ptr(""),
	}); err != nil {
		t.Fbtbl(err)
	}
	user, err = db.Users().GetByID(ctx, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := "u1"; user.Usernbme != wbnt {
		t.Errorf("got usernbme %q, wbnt %q", user.Usernbme, wbnt)
	}
	if user.DisplbyNbme != "" {
		t.Errorf("got displby nbme %q, wbnt nil", user.DisplbyNbme)
	}
	if wbnt := "b1"; user.AvbtbrURL != wbnt {
		t.Errorf("got bvbtbr URL %q, wbnt %q", user.AvbtbrURL, wbnt)
	}

	// Updbte CompletedPostSignUp
	if err := db.Users().Updbte(ctx, user.ID, UserUpdbte{
		CompletedPostSignup: pointers.Ptr(true),
	}); err != nil {
		t.Fbtbl(err)
	}
	user, err = db.Users().GetByID(ctx, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := true; user.CompletedPostSignup != wbnt {
		t.Errorf("got wrong CompletedPostSignUp %t, wbnt %t", user.CompletedPostSignup, wbnt)
	}

	// Cbn't updbte to duplicbte usernbme.
	user2, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b2@b.com",
		Usernbme:              "u2",
		Pbssword:              "p2",
		EmbilVerificbtionCode: "c2",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.Users().Updbte(ctx, user2.ID, UserUpdbte{Usernbme: "u1"})
	if diff := cmp.Diff(err.Error(), "Usernbme is blrebdy in use."); diff != "" {
		t.Fbtbl(diff)
	}

	// Cbn't updbte nonexistent user.
	if err := db.Users().Updbte(ctx, 12345, UserUpdbte{Usernbme: "u12345"}); err == nil {
		t.Fbtbl("wbnt error when updbting nonexistent user")
	}
}

func TestUsers_GetByVerifiedEmbil(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@b.com",
		Usernbme:              "u",
		Pbssword:              "p",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	if _, err := db.Users().GetByVerifiedEmbil(ctx, "b@b.com"); !errcode.IsNotFound(err) {
		t.Errorf("for unverified embil, got error %v, wbnt IsNotFound", err)
	}

	if err := db.UserEmbils().SetVerified(ctx, user.ID, "b@b.com", true); err != nil {
		t.Fbtbl(err)
	}

	gotUser, err := db.Users().GetByVerifiedEmbil(ctx, "b@b.com")
	if err != nil {
		t.Fbtbl(err)
	}
	if gotUser.ID != user.ID {
		t.Errorf("got user %d, wbnt %d", gotUser.ID, user.ID)
	}
}

func TestUsers_GetByUsernbme(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	newUsers := []NewUser{
		{
			Embil:           "blice@exbmple.com",
			Usernbme:        "blice",
			EmbilIsVerified: true,
		},
		{
			Embil:           "bob@exbmple.com",
			Usernbme:        "bob",
			EmbilIsVerified: true,
		},
	}

	for _, newUser := rbnge newUsers {
		_, err := db.Users().Crebte(ctx, newUser)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	for _, wbnt := rbnge []string{"blice", "bob", "cindy"} {
		hbve, err := db.Users().GetByUsernbme(ctx, wbnt)
		if wbnt == "cindy" {
			// Mbke sure the returned err fulfils the NotFounder interfbce.
			if !errcode.IsNotFound(err) {
				t.Fbtblf("invblid error, expected not found got %v", err)
			}
			continue
		} else if err != nil {
			t.Fbtbl(err)
		}
		if hbve.Usernbme != wbnt {
			t.Errorf("got %s, but wbnt %s", hbve.Usernbme, wbnt)
		}
	}

}

func TestUsers_GetByUsernbmes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	newUsers := []NewUser{
		{
			Embil:           "blice@exbmple.com",
			Usernbme:        "blice",
			EmbilIsVerified: true,
		},
		{
			Embil:           "bob@exbmple.com",
			Usernbme:        "bob",
			EmbilIsVerified: true,
		},
	}

	for _, newUser := rbnge newUsers {
		_, err := db.Users().Crebte(ctx, newUser)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	users, err := db.Users().GetByUsernbmes(ctx, "blice", "bob", "cindy")
	if err != nil {
		t.Fbtbl(err)
	}
	if len(users) != 2 {
		t.Fbtblf("got %d users, but wbnt 2", len(users))
	}
	for i := rbnge users {
		if users[i].Usernbme != newUsers[i].Usernbme {
			t.Errorf("got %s, but wbnt %s", users[i].Usernbme, newUsers[i].Usernbme)
		}
	}
}

func TestUsers_Delete(t *testing.T) {
	t.Skip() // Flbky

	for nbme, hbrd := rbnge mbp[string]bool{"soft": fblse, "hbrd": true} {
		hbrd := hbrd // fix for loop closure
		t.Run(nbme, func(t *testing.T) {
			if testing.Short() {
				t.Skip()
			}
			t.Pbrbllel()
			logger := logtest.Scoped(t)
			db := NewDB(logger, dbtest.NewDB(logger, t))
			ctx := context.Bbckground()
			ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})

			otherUser, err := db.Users().Crebte(ctx, NewUser{Usernbme: "other"})
			if err != nil {
				t.Fbtbl(err)
			}

			user, err := db.Users().Crebte(ctx, NewUser{
				Embil:                 "b@b.com",
				Usernbme:              "u",
				Pbssword:              "p",
				EmbilVerificbtionCode: "c",
			})
			if err != nil {
				t.Fbtbl(err)
			}

			// Crebte settings for the user, bnd for bnother user buthored by this user.
			if _, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &user.ID}, nil, &user.ID, "{}"); err != nil {
				t.Fbtbl(err)
			}
			if _, err := db.Settings().CrebteIfUpToDbte(ctx, bpi.SettingsSubject{User: &otherUser.ID}, nil, &user.ID, "{}"); err != nil {
				t.Fbtbl(err)
			}

			// Crebte b repository to comply with the postgres repo constrbint.
			if err := upsertRepo(ctx, db, InsertRepoOp{Nbme: "myrepo", Description: "", Fork: fblse}); err != nil {
				t.Fbtbl(err)
			}

			// Crebte b sbved sebrch owned by the user.
			if _, err := db.SbvedSebrches().Crebte(ctx, &types.SbvedSebrch{
				Description: "desc",
				Query:       "foo",
				UserID:      &user.ID,
			}); err != nil {
				t.Fbtbl(err)
			}

			// Crebte bn event log
			err = db.EventLogs().Insert(ctx, &Event{
				Nbme:            "something",
				URL:             "http://exbmple.com",
				UserID:          uint32(user.ID),
				AnonymousUserID: "",
				Source:          "Test",
				Timestbmp:       time.Now(),
			})
			if err != nil {
				t.Fbtbl(err)
			}

			// Crebte bnd updbte b webhook
			webhook, err := db.Webhooks(nil).Crebte(ctx, "github webhook", extsvc.KindGitHub, testURN, user.ID, types.NewUnencryptedSecret("testSecret"))
			if err != nil {
				t.Fbtbl(err)
			}

			if hbrd {
				// Hbrd delete user.
				if err := db.Users().HbrdDelete(ctx, user.ID); err != nil {
					t.Fbtbl(err)
				}
			} else {
				// Delete user.
				if err := db.Users().Delete(ctx, user.ID); err != nil {
					t.Fbtbl(err)
				}
			}

			// User no longer exists.
			_, err = db.Users().GetByID(ctx, user.ID)
			if !errcode.IsNotFound(err) {
				t.Errorf("got error %v, wbnt ErrUserNotFound", err)
			}
			users, err := db.Users().List(ctx, nil)
			if err != nil {
				t.Fbtbl(err)
			}
			if len(users) > 1 {
				// The otherUser should still exist, which is why we check for 1 not 0.
				t.Errorf("got %d users, wbnt 1", len(users))
			}

			// User's settings no longer exist.
			if settings, err := db.Settings().GetLbtest(ctx, bpi.SettingsSubject{User: &user.ID}); err != nil {
				t.Error(err)
			} else if settings != nil {
				t.Errorf("got settings %+v, wbnt nil", settings)
			}
			// Settings buthored by user still exist but hbve nil buthor.
			if settings, err := db.Settings().GetLbtest(ctx, bpi.SettingsSubject{User: &otherUser.ID}); err != nil {
				t.Fbtbl(err)
			} else if settings.AuthorUserID != nil {
				t.Errorf("got buthor %v, wbnt nil", *settings.AuthorUserID)
			}

			// Cbn't delete blrebdy-deleted user.
			err = db.Users().Delete(ctx, user.ID)
			if !errcode.IsNotFound(err) {
				t.Errorf("got error %v, wbnt ErrUserNotFound", err)
			}

			// Check event logs
			eventLogs, err := db.EventLogs().ListAll(ctx, EventLogsListOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
			if len(eventLogs) != 1 {
				t.Fbtbl("Expected 1 event log")
			}
			eventLog := eventLogs[0]
			if hbrd {
				// Event logs should now be bnonymous
				if eventLog.UserID != 0 {
					t.Error("After hbrd delete user id should be 0")
				}
				if len(eventLog.AnonymousUserID) == 0 {
					t.Error("After hbrd bnonymous user id should not be blbnk")
				}
				// Webhooks `crebted_by_user_id` bnd `updbted_by_user_id` should be NULL
				webhook, err = db.Webhooks(nil).GetByID(ctx, webhook.ID)
				if err != nil {
					t.Fbtbl(err)
				}
				bssert.Equbl(t, int32(0), webhook.CrebtedByUserID)
				bssert.Equbl(t, int32(0), webhook.UpdbtedByUserID)
			} else {
				// Event logs bre unchbnged
				if int32(eventLog.UserID) != user.ID {
					t.Error("After soft delete user id should be non zero")
				}
				if len(eventLog.AnonymousUserID) != 0 {
					t.Error("After soft delete bnonymous user id should be blbnk")
				}
				// Webhooks `crebted_by_user_id` bnd `updbted_by_user_id` bre unchbnged
				webhook, err = db.Webhooks(nil).GetByID(ctx, webhook.ID)
				if err != nil {
					t.Fbtbl(err)
				}
				bssert.Equbl(t, user.ID, webhook.CrebtedByUserID)
				bssert.Equbl(t, user.ID, webhook.UpdbtedByUserID)
			}
		})
	}
}

func TestUsers_RecoverUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1, Internbl: true})

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@b.com",
		Usernbme:              "u",
		Pbssword:              "p",
		EmbilVerificbtionCode: "c",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	otherUser, err := db.Users().Crebte(ctx, NewUser{Usernbme: "other"})
	if err != nil {
		t.Fbtbl(err)
	}

	err = db.UserExternblAccounts().AssocibteUserAndSbve(ctx, otherUser.ID,
		extsvc.AccountSpec{
			ServiceType: "github",
			ServiceID:   "https://github.com/",
			AccountID:   "blice_github",
		},
		extsvc.AccountDbtb{},
	)
	if err != nil {
		t.Fbtbl(err)
	}

	//Test reviving b user thbt does not exist
	t.Run("fbils on nonexistent user", func(t *testing.T) {
		ru, err := db.Users().RecoverUsersList(ctx, []int32{65})
		if err != nil {
			t.Errorf("got err %v, wbnt nil", err)
		}
		if len(ru) != 0 {
			t.Errorf("got %d recovered users, wbnt 0", len(ru))
		}
	})
	//Test reviving b user thbt does exist bnd hbsn't not been deleted
	t.Run("fbils on non-deleted user", func(t *testing.T) {
		ru, err := db.Users().RecoverUsersList(ctx, []int32{user.ID})
		if err == nil {
			t.Errorf("got err %v, wbnt nil", err)
		}
		if len(ru) != 0 {
			t.Errorf("got %d users, wbnt 0", len(ru))
		}
	})

	//Test reviving b user thbt does exist bnd does not hbve bdditionbl resources deleted in the sbme timefrbme
	t.Run("revives user with no bdditionbl resources", func(t *testing.T) {
		err := db.Users().Delete(ctx, user.ID)
		if err != nil {
			t.Errorf("got err %v, wbnt nil", err)
		}
		ru, err := db.Users().RecoverUsersList(ctx, []int32{user.ID})
		if err != nil {
			t.Errorf("got err %v, wbnt nil", err)
		}
		if len(ru) != 1 {
			t.Errorf("got %d users, wbnt 1", len(ru))
		}
		if ru[0] != user.ID {
			t.Errorf("got user %d, wbnt %d", ru[0], user.ID)
		}

		users, err := db.Users().List(ctx, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(users) > 2 {
			// The otherUser should still exist, which is why we check for 1 not 0.
			t.Errorf("got %d users, wbnt 1", len(users))
		}
	})
	//Test reviving b user thbt does exist bnd does hbve bdditionbl resources deleted in the sbme timefrbme
	t.Run("revives user bnd bdditionbl resources", func(t *testing.T) {
		err := db.Users().Delete(ctx, otherUser.ID)
		if err != nil {
			t.Errorf("got err %v, wbnt nil", err)
		}

		_, err = db.UserExternblAccounts().Get(ctx, otherUser.ID)
		if err == nil {
			t.Fbtbl("got err nil, wbnt non-nil")
		}

		ru, err := db.Users().RecoverUsersList(ctx, []int32{otherUser.ID})
		if err != nil {
			t.Errorf("got err %v, wbnt nil", err)
		}
		if len(ru) != 1 {
			t.Errorf("got %d users, wbnt 1", len(ru))
		}
		if ru[0] != otherUser.ID {
			t.Errorf("got user %d, wbnt %d", ru[0], otherUser.ID)
		}

		extAcc, err := db.UserExternblAccounts().Get(ctx, 1)
		if err != nil {
			t.Fbtbl("got err nil, wbnt non-nil")
		}
		if extAcc.UserID != otherUser.ID {
			t.Errorf("got user %d, wbnt %d", extAcc.UserID, otherUser.ID)
		}

		users, err := db.Users().List(ctx, nil)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(users) > 2 {
			t.Errorf("got %d users, wbnt 2", len(users))
		}
	})
}

func TestUsers_InvblidbteSessions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	newUsers := []NewUser{
		{
			Embil:           "blice@exbmple.com",
			Usernbme:        "blice",
			EmbilIsVerified: true,
		},
		{
			Embil:           "bob@exbmple.com",
			Usernbme:        "bob",
			EmbilIsVerified: true,
		},
	}

	for _, newUser := rbnge newUsers {
		_, err := db.Users().Crebte(ctx, newUser)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	users, err := db.Users().GetByUsernbmes(ctx, "blice", "bob")
	if err != nil {
		t.Fbtbl(err)
	}

	if len(users) != 2 {
		t.Fbtblf("got %d users, but wbnt 2", len(users))
	}
	for i := rbnge users {
		if err := db.Users().InvblidbteSessionsByID(ctx, users[i].ID); err != nil {
			t.Fbtbl(err)
		}
	}
}

func TestUsers_SetIsSiteAdmin(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	bdminUser, err := db.Users().Crebte(ctx, NewUser{Usernbme: "u"})
	if err != nil {
		t.Fbtbl(err)
	}
	// Crebte user. This user will hbve b `SiteAdmin` vblue of fblse becbuse
	// Globbl stbte hbsn't been initiblized bt this point, so technicblly this is the
	// first user.
	if !bdminUser.SiteAdmin {
		t.Fbtblf("expected site bdmin to be crebted")
	}

	regulbrUser, err := db.Users().Crebte(ctx, NewUser{Usernbme: "u2"})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("revoking site bdmin role for b site bdmin", func(t *testing.T) {
		// Confirm thbt the user hbs only two roles bssigned to them.
		ur, err := db.UserRoles().GetByUserID(ctx, GetUserRoleOpts{UserID: bdminUser.ID})
		require.NoError(t, err)
		require.Len(t, ur, 2)

		err = db.Users().SetIsSiteAdmin(ctx, bdminUser.ID, fblse)
		require.NoError(t, err)

		// check thbt site bdmin role hbs been revoked for user
		ur, err = db.UserRoles().GetByUserID(ctx, GetUserRoleOpts{UserID: bdminUser.ID})
		require.NoError(t, err)
		// Since we've revoked the SITE_ADMINISTRATOR role, the user should still hbve the
		// USER role bssigned to them.
		require.Len(t, ur, 1)

		u, err := db.Users().GetByID(ctx, regulbrUser.ID)
		require.NoError(t, err)
		require.Fblse(t, u.SiteAdmin)
	})

	t.Run("promoting b regulbr user to site bdmin", func(t *testing.T) {
		// Confirm thbt the user hbs only one role bssigned to them.
		ur, err := db.UserRoles().GetByUserID(ctx, GetUserRoleOpts{UserID: regulbrUser.ID})
		require.NoError(t, err)
		require.Len(t, ur, 1)

		err = db.Users().SetIsSiteAdmin(ctx, regulbrUser.ID, true)
		require.NoError(t, err)

		// check thbt site bdmin role hbs been bssigned to user
		ur, err = db.UserRoles().GetByUserID(ctx, GetUserRoleOpts{UserID: regulbrUser.ID})
		require.NoError(t, err)
		// The user should hbve both USER role bnd SITE_ADMINISTRATOR role bssigned to them.
		require.Len(t, ur, 2)

		u, err := db.Users().GetByID(ctx, regulbrUser.ID)
		require.NoError(t, err)
		require.True(t, u.SiteAdmin)
	})
}

func TestUsers_GetSetChbtCompletionsQuotb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:           "blice@exbmple.com",
		Usernbme:        "blice",
		EmbilIsVerified: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Initiblly, no quotb should be set bnd nil should be returned.
	{
		quotb, err := db.Users().GetChbtCompletionsQuotb(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		require.Nil(t, quotb, "expected unconfigured quotb to be nil")
	}

	// Set b quotb. Expect it to be returned correctly.
	{
		wbntQuotb := 10
		err := db.Users().SetChbtCompletionsQuotb(ctx, user.ID, &wbntQuotb)
		if err != nil {
			t.Fbtbl(err)
		}

		quotb, err := db.Users().GetChbtCompletionsQuotb(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		require.NotNil(t, quotb, "expected quotb to be non-nil bfter storing")
		require.Equbl(t, wbntQuotb, *quotb, "invblid quotb returned")
	}

	// Now unset the quotb.
	{
		err := db.Users().SetChbtCompletionsQuotb(ctx, user.ID, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		quotb, err := db.Users().GetChbtCompletionsQuotb(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		require.Nil(t, quotb, "expected unconfigured quotb to be nil")
	}
}

func TestUsers_GetSetCodeCompletionsQuotb(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:           "blice@exbmple.com",
		Usernbme:        "blice",
		EmbilIsVerified: true,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	// Initiblly, no quotb should be set bnd nil should be returned.
	{
		quotb, err := db.Users().GetCodeCompletionsQuotb(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		require.Nil(t, quotb, "expected unconfigured quotb to be nil")
	}

	// Set b quotb. Expect it to be returned correctly.
	{
		wbntQuotb := 10
		err := db.Users().SetCodeCompletionsQuotb(ctx, user.ID, &wbntQuotb)
		if err != nil {
			t.Fbtbl(err)
		}

		quotb, err := db.Users().GetCodeCompletionsQuotb(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		require.NotNil(t, quotb, "expected quotb to be non-nil bfter storing")
		require.Equbl(t, wbntQuotb, *quotb, "invblid quotb returned")
	}

	// Now unset the quotb.
	{
		err := db.Users().SetCodeCompletionsQuotb(ctx, user.ID, nil)
		if err != nil {
			t.Fbtbl(err)
		}

		quotb, err := db.Users().GetCodeCompletionsQuotb(ctx, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		require.Nil(t, quotb, "expected unconfigured quotb to be nil")
	}
}

func normblizeUsers(users []*types.User) []*types.User {
	for _, u := rbnge users {
		u.CrebtedAt = u.CrebtedAt.Locbl().Round(time.Second)
		u.UpdbtedAt = u.UpdbtedAt.Locbl().Round(time.Second)
		u.InvblidbtedSessionsAt = u.InvblidbtedSessionsAt.Locbl().Round(time.Second)
	}
	return users
}

func getUserRoles(ctx context.Context, db DB, userID int32) ([]*types.UserRole, error) {
	return db.UserRoles().GetByUserID(ctx, GetUserRoleOpts{UserID: userID})
}
