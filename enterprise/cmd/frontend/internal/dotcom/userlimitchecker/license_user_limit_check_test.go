package userlimitchecker

/* import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log"
	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/stretchr/testify/require"
)

func TestCreateUserLimitChecker(t *testing.T) {
	logger := log.NoOp()
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	// create user to satisfy product_subscription foreign key constraint
	userStore := db.Users()
	user, err := userStore.Create(ctx, database.NewUser{
		Email:                 "marty@mcfly.com",
		Username:              "martymcfly",
		DisplayName:           "marty",
		Password:              "docbrown",
		AvatarURL:             "calvinkleinundies.jpg",
		EmailVerificationCode: "51235",
		EmailIsVerified:       true,
		FailIfNotInitialUser:  false,
		EnforcePasswordLength: false,
		TosAccepted:           false,
	})
	require.NoError(t, err)

	// create product_subscription to satisfy product_license foreign key constraint
	subStore := ps.NewDbSubscription(db)
	subId, err := subStore.Create(ctx, user.ID, user.Username)
	require.NoError(t, err)

	// license can now be created
	licensesStore := ps.NewDbLicense(db)
	id, err := licensesStore.Create(ctx, subId, "12345", 5, license.Info{
		UserCount: 2,
		ExpiresAt: time.Now().Add(14 * 24 * time.Hour),
	})
	require.NoError(t, err)
} */
