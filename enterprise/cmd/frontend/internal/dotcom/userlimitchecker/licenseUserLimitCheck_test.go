package userlimitchecker

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log"
	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
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
		AvatarURL:             "calvinkleins.jpg",
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

	// create license
	licensesStore := ps.NewDbLicense(db)
	licenseId, err := licensesStore.Create(ctx, subId, "12345", 5, license.Info{
		UserCount: 2,
		ExpiresAt: time.Now().Add(14 * 24 * time.Hour),
	})
	require.NoError(t, err)

	checkerStore := NewUserLimitChecker(db)
	checkerId, err := checkerStore.Create(ctx, licenseId, 20)
	require.NoError(t, err)
	assert.NotNil(t, checkerId)
}

func TestGetByLicenseID(t *testing.T) {
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
		AvatarURL:             "calvinkleins.jpg",
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

	// create license
	licensesStore := ps.NewDbLicense(db)
	licenseId, err := licensesStore.Create(ctx, subId, "12345", 5, license.Info{
		UserCount: 20,
		ExpiresAt: time.Now().Add(14 * 24 * time.Hour),
	})
	require.NoError(t, err)

	checkerStore := NewUserLimitChecker(db)
	checkerId, err := checkerStore.Create(ctx, licenseId, 20)
	require.NoError(t, err)

	checker, err := checkerStore.GetByLicenseID(ctx, licenseId)
	require.NoError(t, err)

	assert.Equal(t, checker.ID, checkerId)
	assert.Equal(t, checker.LicenseID, licenseId)
	assert.Nil(t, checker.UserCountAlertSentAt)
	assert.Equal(t, 20, checker.UserCountWhenEmailLastSent)
	assert.NotZero(t, checker.CreatedAt)
	assert.NotZero(t, checker.UpdatedAt)
}
