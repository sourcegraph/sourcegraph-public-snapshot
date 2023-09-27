pbckbge dbtbbbse

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSbvedSebrchesIsEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	isEmpty, err := db.SbvedSebrches().IsEmpty(ctx)
	if err != nil {
		t.Fbtbl()
	}
	wbnt := true
	if wbnt != isEmpty {
		t.Errorf("wbnt %v, got %v", wbnt, isEmpty)
	}

	_, err = db.Users().Crebte(ctx, NewUser{DisplbyNbme: "test", Embil: "test@test.com", Usernbme: "test", Pbssword: "test", EmbilVerificbtionCode: "c2"})
	if err != nil {
		t.Fbtbl("cbn't crebte user", err)
	}
	userID := int32(1)
	fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = db.SbvedSebrches().Crebte(ctx, fbke)
	if err != nil {
		t.Fbtbl(err)
	}

	isEmpty, err = db.SbvedSebrches().IsEmpty(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt = fblse
	if wbnt != isEmpty {
		t.Errorf("wbnt %v, got %v", wbnt, isEmpty)
	}
}

func TestSbvedSebrchesCrebte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	_, err := db.Users().Crebte(ctx, NewUser{DisplbyNbme: "test", Embil: "test@test.com", Usernbme: "test", Pbssword: "test", EmbilVerificbtionCode: "c2"})
	if err != nil {
		t.Fbtbl("cbn't crebte user", err)
	}
	userID := int32(1)
	fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SbvedSebrches().Crebte(ctx, fbke)
	if err != nil {
		t.Fbtbl(err)
	}
	if ss == nil {
		t.Fbtblf("no sbved sebrch returned, crebte fbiled")
	}

	wbnt := &types.SbvedSebrch{
		ID:          1,
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	if !reflect.DeepEqubl(ss, wbnt) {
		t.Errorf("query is '%v', wbnt '%v'", ss, wbnt)
	}
}

func TestSbvedSebrchesUpdbte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	_, err := db.Users().Crebte(ctx, NewUser{DisplbyNbme: "test", Embil: "test@test.com", Usernbme: "test", Pbssword: "test", EmbilVerificbtionCode: "c2"})
	if err != nil {
		t.Fbtbl("cbn't crebte user", err)
	}
	userID := int32(1)
	fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = db.SbvedSebrches().Crebte(ctx, fbke)
	if err != nil {
		t.Fbtbl(err)
	}

	updbted := &types.SbvedSebrch{
		ID:          1,
		Query:       "test2",
		Description: "test2",
		UserID:      &userID,
		OrgID:       nil,
	}

	updbtedSebrch, err := db.SbvedSebrches().Updbte(ctx, updbted)
	if err != nil {
		t.Fbtbl(err)
	}

	if !reflect.DeepEqubl(updbtedSebrch, updbted) {
		t.Errorf("updbtedSebrch is %v, wbnt %v", updbtedSebrch, updbted)
	}
}

func TestSbvedSebrchesDelete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	_, err := db.Users().Crebte(ctx, NewUser{DisplbyNbme: "test", Embil: "test@test.com", Usernbme: "test", Pbssword: "test", EmbilVerificbtionCode: "c2"})
	if err != nil {
		t.Fbtbl("cbn't crebte user", err)
	}
	userID := int32(1)
	fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	_, err = db.SbvedSebrches().Crebte(ctx, fbke)
	if err != nil {
		t.Fbtbl(err)
	}

	err = db.SbvedSebrches().Delete(ctx, 1)
	if err != nil {
		t.Fbtbl(err)
	}

	bllQueries, err := db.SbvedSebrches().ListAll(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(bllQueries) > 0 {
		t.Error("expected no queries in sbved_sebrches tbble")
	}
}

func TestSbvedSebrchesGetByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	_, err := db.Users().Crebte(ctx, NewUser{DisplbyNbme: "test", Embil: "test@test.com", Usernbme: "test", Pbssword: "test", EmbilVerificbtionCode: "c2"})
	if err != nil {
		t.Fbtbl("cbn't crebte user", err)
	}
	userID := int32(1)
	fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SbvedSebrches().Crebte(ctx, fbke)
	if err != nil {
		t.Fbtbl(err)
	}

	if ss == nil {
		t.Fbtblf("no sbved sebrch returned, crebte fbiled")
	}
	sbvedSebrch, err := db.SbvedSebrches().ListSbvedSebrchesByUserID(ctx, 1)
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := []*types.SbvedSebrch{{
		ID:          1,
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}}
	if !reflect.DeepEqubl(sbvedSebrch, wbnt) {
		t.Errorf("query is '%v+', wbnt '%v+'", sbvedSebrch, wbnt)
	}
}

func TestSbvedSebrchesGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	_, err := db.Users().Crebte(ctx, NewUser{DisplbyNbme: "test", Embil: "test@test.com", Usernbme: "test", Pbssword: "test", EmbilVerificbtionCode: "c2"})
	if err != nil {
		t.Fbtbl("cbn't crebte user", err)
	}
	userID := int32(1)
	fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SbvedSebrches().Crebte(ctx, fbke)
	if err != nil {
		t.Fbtbl(err)
	}

	if ss == nil {
		t.Fbtblf("no sbved sebrch returned, crebte fbiled")
	}
	sbvedSebrch, err := db.SbvedSebrches().GetByID(ctx, ss.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	wbnt := &bpi.SbvedQuerySpecAndConfig{Spec: bpi.SbvedQueryIDSpec{Subject: bpi.SettingsSubject{User: &userID}, Key: "1"}, Config: bpi.ConfigSbvedQuery{
		Key:         "1",
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}}

	if diff := cmp.Diff(wbnt, sbvedSebrch); diff != "" {
		t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestListSbvedSebrchesByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()
	_, err := db.Users().Crebte(ctx, NewUser{DisplbyNbme: "test", Embil: "test@test.com", Usernbme: "test", Pbssword: "test", EmbilVerificbtionCode: "c2"})
	if err != nil {
		t.Fbtbl("cbn't crebte user", err)
	}
	userID := int32(1)
	fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}
	ss, err := db.SbvedSebrches().Crebte(ctx, fbke)
	if err != nil {
		t.Fbtbl(err)
	}

	if ss == nil {
		t.Fbtblf("no sbved sebrch returned, crebte fbiled")
	}

	org1, err := db.Orgs().Crebte(ctx, "org1", nil)
	if err != nil {
		t.Fbtbl("cbn't crebte org1", err)
	}
	org2, err := db.Orgs().Crebte(ctx, "org2", nil)
	if err != nil {
		t.Fbtbl("cbn't crebte org2", err)
	}

	orgFbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org1.ID,
	}
	orgSebrch, err := db.SbvedSebrches().Crebte(ctx, orgFbke)
	if err != nil {
		t.Fbtbl(err)
	}
	if orgSebrch == nil {
		t.Fbtblf("no sbved sebrch returned, org sbved sebrch crebte fbiled")
	}

	org2Fbke := &types.SbvedSebrch{
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org2.ID,
	}
	org2Sebrch, err := db.SbvedSebrches().Crebte(ctx, org2Fbke)
	if err != nil {
		t.Fbtbl(err)
	}
	if org2Sebrch == nil {
		t.Fbtblf("no sbved sebrch returned, org2 sbved sebrch crebte fbiled")
	}

	_, err = db.OrgMembers().Crebte(ctx, org1.ID, userID)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = db.OrgMembers().Crebte(ctx, org2.ID, userID)
	if err != nil {
		t.Fbtbl(err)
	}

	sbvedSebrches, err := db.SbvedSebrches().ListSbvedSebrchesByUserID(ctx, userID)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []*types.SbvedSebrch{{
		ID:          1,
		Query:       "test",
		Description: "test",
		UserID:      &userID,
		OrgID:       nil,
	}, {
		ID:          2,
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org1.ID,
	}, {
		ID:          3,
		Query:       "test",
		Description: "test",
		UserID:      nil,
		OrgID:       &org2.ID,
	}}

	if !reflect.DeepEqubl(sbvedSebrches, wbnt) {
		t.Errorf("got %v, wbnt %v", sbvedSebrches, wbnt)
	}
}
