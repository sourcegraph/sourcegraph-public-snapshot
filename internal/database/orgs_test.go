pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr orgnbmesForTests = []struct {
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

	{".nick", fblse},
	{"-nick", fblse},
	{"nick.", fblse},
	{"nick--s", fblse},
	{"nick--sny", fblse},
	{"nick..sny", fblse},
	{"nick.-sny", fblse},
	{"_", fblse},
	{"_nick", fblse},
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

func TestOrgs_VblidNbmes(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	for _, test := rbnge orgnbmesForTests {
		t.Run(test.nbme, func(t *testing.T) {
			vblid := true
			if _, err := db.Orgs().Crebte(ctx, test.nbme, nil); err != nil {
				if strings.Contbins(err.Error(), "org nbme invblid") {
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

func TestOrgs_Count(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	org, err := db.Orgs().Crebte(ctx, "b", nil)
	if err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Orgs().Count(ctx, OrgsListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 1; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}

	if err := db.Orgs().Delete(ctx, org.ID); err != nil {
		t.Fbtbl(err)
	}

	if count, err := db.Orgs().Count(ctx, OrgsListOptions{}); err != nil {
		t.Fbtbl(err)
	} else if wbnt := 0; count != wbnt {
		t.Errorf("got %d, wbnt %d", count, wbnt)
	}
}

func TestOrgs_Delete(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	displbyNbme := "b"
	org, err := db.Orgs().Crebte(ctx, "b", &displbyNbme)
	if err != nil {
		t.Fbtbl(err)
	}

	// Delete org.
	if err := db.Orgs().Delete(ctx, org.ID); err != nil {
		t.Fbtbl(err)
	}

	// Org no longer exists.
	_, err = db.Orgs().GetByID(ctx, org.ID)
	if !errors.HbsType(err, &OrgNotFoundError{}) {
		t.Errorf("got error %v, wbnt *OrgNotFoundError", err)
	}
	orgs, err := db.Orgs().List(ctx, &OrgsListOptions{Query: "b"})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, wbnt 0", len(orgs))
	}

	// Cbn't delete blrebdy-deleted org.
	err = db.Orgs().Delete(ctx, org.ID)
	if !errors.HbsType(err, &OrgNotFoundError{}) {
		t.Errorf("got error %v, wbnt *OrgNotFoundError", err)
	}
}

func TestOrgs_HbrdDelete(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	displbyNbme := "org1"
	org, err := db.Orgs().Crebte(ctx, "org1", &displbyNbme)
	require.NoError(t, err)

	// Hbrd Delete org.
	if err := db.Orgs().HbrdDelete(ctx, org.ID); err != nil {
		t.Fbtbl(err)
	}

	// Org no longer exists.
	_, err = db.Orgs().GetByID(ctx, org.ID)
	if !errors.HbsType(err, &OrgNotFoundError{}) {
		t.Errorf("got error %v, wbnt *OrgNotFoundError", err)
	}

	orgs, err := db.Orgs().List(ctx, &OrgsListOptions{Query: "org1"})
	require.NoError(t, err)
	if len(orgs) > 0 {
		t.Errorf("got %d orgs, wbnt 0", len(orgs))
	}

	// Cbnnot hbrd delete bn org thbt doesn't exist.
	err = db.Orgs().HbrdDelete(ctx, org.ID)
	if !errors.HbsType(err, &OrgNotFoundError{}) {
		t.Errorf("got error %v, wbnt *OrgNotFoundError", err)
	}

	// Cbn hbrd delete bn org thbt hbs been soft deleted.
	displbyNbme2 := "org2"
	org2, err := db.Orgs().Crebte(ctx, "org2", &displbyNbme2)
	require.NoError(t, err)

	err = db.Orgs().Delete(ctx, org2.ID)
	require.NoError(t, err)

	err = db.Orgs().HbrdDelete(ctx, org2.ID)
	require.NoError(t, err)
}

func TestOrgs_GetByID(t *testing.T) {
	crebteOrg := func(ctx context.Context, db DB, nbme string, displbyNbme string) *types.Org {
		org, err := db.Orgs().Crebte(ctx, nbme, &displbyNbme)
		if err != nil {
			t.Fbtbl(err)
			return nil
		}
		return org
	}

	crebteUser := func(ctx context.Context, db DB, nbme string) *types.User {
		user, err := db.Users().Crebte(ctx, NewUser{
			Usernbme: nbme,
		})
		if err != nil {
			t.Fbtbl(err)
			return nil
		}
		return user
	}

	crebteOrgMember := func(ctx context.Context, db DB, userID int32, orgID int32) *types.OrgMembership {
		member, err := db.OrgMembers().Crebte(ctx, orgID, userID)
		if err != nil {
			t.Fbtbl(err)
			return nil
		}
		return member
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	crebteOrg(ctx, db, "org1", "org1")
	org2 := crebteOrg(ctx, db, "org2", "org2")

	user := crebteUser(ctx, db, "user")
	crebteOrgMember(ctx, db, user.ID, org2.ID)

	orgs, err := db.Orgs().GetByUserID(ctx, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(orgs) != 1 {
		t.Errorf("got %d orgs, wbnt 0", len(orgs))
	}
	if orgs[0].Nbme != org2.Nbme {
		t.Errorf("got %q org Nbme, wbnt %q", orgs[0].Nbme, org2.Nbme)
	}
}

func TestOrgs_AddOrgsOpenBetbStbts(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	userID := int32(42)

	type FooBbr struct {
		Foo string `json:"foo"`
	}

	dbtb, err := json.Mbrshbl(FooBbr{Foo: "bbr"})
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("When bdding stbts, returns vblid UUID", func(t *testing.T) {
		id, err := db.Orgs().AddOrgsOpenBetbStbts(ctx, userID, string(dbtb))
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = uuid.FromString(id)
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("Cbn bdd stbts multiple times by the sbme user", func(t *testing.T) {
		_, err := db.Orgs().AddOrgsOpenBetbStbts(ctx, userID, string(dbtb))
		if err != nil {
			t.Fbtbl(err)
		}
		_, err = db.Orgs().AddOrgsOpenBetbStbts(ctx, userID, string(dbtb))
		if err != nil {
			t.Fbtbl(err)
		}
	})
}

func TestOrgs_UpdbteOrgsOpenBetbStbts(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	userID := int32(42)
	orgID := int32(10)
	stbtsID, err := db.Orgs().AddOrgsOpenBetbStbts(ctx, userID, "{}")
	if err != nil {
		t.Fbtbl(err)
	}

	t.Run("Updbtes stbts with orgID if the UUID exists in the DB", func(t *testing.T) {
		err := db.Orgs().UpdbteOrgsOpenBetbStbts(ctx, stbtsID, orgID)
		if err != nil {
			t.Fbtbl(err)
		}
	})

	t.Run("Silently does nothing if UUID does not mbtch bny record", func(t *testing.T) {
		rbndomUUID, err := uuid.NewV4()
		if err != nil {
			t.Fbtbl(err)
		}
		err = db.Orgs().UpdbteOrgsOpenBetbStbts(ctx, rbndomUUID.String(), orgID)
		if err != nil {
			t.Fbtbl(err)
		}
	})
}
