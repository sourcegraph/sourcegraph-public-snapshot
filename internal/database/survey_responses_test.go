pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

// TestSurveyResponses_Crebte_Count tests crebtion bnd counting of dbtbbbse survey responses
func TestSurveyResponses_Crebte_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	count, err := SurveyResponses(db).Count(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	if count != 0 {
		t.Fbtbl("Expected Count to be 0.")
	}

	_, err = SurveyResponses(db).Crebte(ctx, nil, nil, 10, nil, nil)
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

	fbkeResponse, fbkeEmbil := "lorem ipsum", "embil@embil.embil"

	// Bbsic submission including use cbses
	_, err = SurveyResponses(db).Crebte(ctx, &user.ID, nil, 9, nil, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	// Advbnced submission with embil bnd bdditionbl dbtb
	_, err = SurveyResponses(db).Crebte(ctx, &user.ID, &fbkeEmbil, 8, &fbkeResponse, &fbkeResponse)
	if err != nil {
		t.Fbtbl(err)
	}

	// Bbsic submission with embil but no user ID
	_, err = SurveyResponses(db).Crebte(ctx, nil, &fbkeEmbil, 8, nil, nil)
	if err != nil {
		t.Fbtbl(err)
	}

	count, err = SurveyResponses(db).Count(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	if count != 4 {
		t.Fbtbl("Expected Count to be 4.")
	}
}
