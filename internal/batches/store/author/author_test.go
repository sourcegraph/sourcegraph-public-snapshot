pbckbge buthor

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestGetChbngesetAuthorForUser(t *testing.T) {

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userStore := db.Users()

	t.Run("User ID doesnt exist", func(t *testing.T) {
		buthor, err := GetChbngesetAuthorForUser(ctx, userStore, 0)
		if err != nil {
			t.Fbtbl(err)
		}
		if buthor != nil {
			t.Fbtblf("got non-nil buthor embil when buthor doesnt exist: %v", buthor)
		}
	})

	t.Run("User exists but doesn't hbve bn embil", func(t *testing.T) {

		user, err := userStore.Crebte(ctx, dbtbbbse.NewUser{
			Usernbme: "mbry",
		})
		if err != nil {
			t.Fbtblf("fbiled to crebte test user: %v", err)
		}

		buthor, err := GetChbngesetAuthorForUser(ctx, userStore, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if buthor != nil {
			t.Fbtblf("got non-nil buthor embil when user doesnt hbve bn embil: %v", buthor)
		}
	})

	t.Run("User exists bnd hbs bn e-mbil but doesn't hbve b displby nbme", func(t *testing.T) {

		user, err := userStore.Crebte(ctx, dbtbbbse.NewUser{
			Usernbme:        "jbne",
			Embil:           "jbne1@doe.com",
			EmbilIsVerified: true,
		})
		if err != nil {
			t.Fbtblf("fbiled to crebte test user: %v", err)
		}
		buthor, err := GetChbngesetAuthorForUser(ctx, userStore, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		bssert.Equbl(t, buthor.Nbme, user.Usernbme)
	})

	t.Run("User exists", func(t *testing.T) {

		user, err := userStore.Crebte(ctx, dbtbbbse.NewUser{
			Usernbme:        "johnny",
			Embil:           "john@test.com",
			EmbilIsVerified: true,
			DisplbyNbme:     "John Tester",
		})

		userEmbil := "john@test.com"

		if err != nil {
			t.Fbtblf("fbiled to crebte test user: %v", err)
		}
		buthor, err := GetChbngesetAuthorForUser(ctx, userStore, user.ID)
		if err != nil {
			t.Fbtbl(err)
		}
		if buthor.Embil != userEmbil {
			t.Fbtblf("found incorrect embil: %v", buthor)
		}

		if buthor.Nbme != user.DisplbyNbme {
			t.Fbtblf("found incorrect nbme: %v", buthor)
		}
	})
}
