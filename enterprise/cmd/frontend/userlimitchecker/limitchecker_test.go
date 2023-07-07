package userlimitchecker

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/log"
	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestAtOrOverUserLimit(t *testing.T) {
	ctx := context.Background()
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	userStore := db.Users()

	var createdUsers []*types.User
	for i, user := range users {
		newUser, err := userStore.Create(ctx, user)
		if err != nil {
			t.Errorf("could not create new user %s", err)
		}
		createdUsers = append(createdUsers, newUser)
		if i == 0 {
			createdUsers[i].SiteAdmin = true
		}
	}

	got, _ := atOrOverUserLimit(ctx, db)
	want := true
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

// for product_licenses
// - product_subscription maps to product_subscription(id)
// for product_subscription
// - user_id maps to users(id)

// steps to resolve this test:
// create a user.
// create a product_subscription with user_id set to the created user(id)
// create a product_license with the product_subscription_id set to created product_subscription(id)

func TestGetLicenseUserLimit(t *testing.T) {
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()

	// product subscription has foreign key of user_id, so we create a user.
	userStore := db.Users()
	user, err := userStore.Create(ctx, users[0])
	if err != nil {
		t.Errorf("could not create user: %s", err)
	}

	// product_license has foreign key of product_subscription_id, so we create a product_subscription.
	subStore := ps.NewDbSubscription(db)
	sub, err := subStore.Create(ctx, user.ID, user.Username)
	if err != nil {
		t.Errorf("could not create subscription: %s", err)
	}

	// use product_subscription_id when creating the license
	licensesStore := ps.NewDbLicense(db)
	for _, license := range licensesToCreate {
		_, err = licensesStore.Create(
			ctx,
			sub,
			license.licenseId,
			license.version,
			license.licenseInfo,
		)
		if err != nil {
			t.Errorf("could not create license:, %s", err)
		}
	}

	got, err := getLicenseUserLimit(ctx, db)
	if err != nil {
		t.Errorf("could not get user limit: %s", err)
	}

	want := 30
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}

func TestGetUserCount(t *testing.T) {
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	userStore := db.Users()

	var createdUsers []*types.User
	for i, user := range users {
		newUser, err := userStore.Create(ctx, user)
		if err != nil {
			t.Errorf("could not create new user %s", err)
		}
		createdUsers = append(createdUsers, newUser)
		if i == 0 {
			createdUsers[i].SiteAdmin = true
		}
	}

	got, err := getUserCount(ctx, db)
	if err != nil {
		t.Errorf("could not get user count: %s", err)
	}

	want := 4
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
}

func TestGetSiteAdmins(t *testing.T) {
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	userStore := db.Users()

	var createdUsers []*types.User
	for i, user := range users {
		newUser, err := userStore.Create(ctx, user)
		if err != nil {
			t.Errorf("could not create new user %s", err)
		}
		createdUsers = append(createdUsers, newUser)

		// set users 1 and 3 to admins
		if i == 0 || i == 2 {
			userStore.SetIsSiteAdmin(ctx, createdUsers[i].ID, true)
		}
	}

	got, _ := getSiteAdmins(ctx, db)
	want := []string{"test@test.com", "test3@test.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestGetUserEmail(t *testing.T) {
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	userStore := db.Users()

	cases := []struct {
		want string
		user database.NewUser
	}{
		{
			want: "test@test.com",
			user: users[0],
		},
		{
			want: "test2@test.com",
			user: users[1],
		},
		{
			want: "test3@test.com",
			user: users[2],
		},
		{
			want: "test4@test.com",
			user: users[3],
		},
	}

	for _, tc := range cases {
		newUser, err := userStore.Create(ctx, tc.user)
		if err != nil {
			t.Errorf("could not create new user: %s", err)
		}

		got, _, err := getUserEmail(ctx, db, newUser)
		if err != nil {
			t.Errorf("got an unexpected error: %s", err)
		}

		if got != tc.want {
			t.Errorf("got %s want %s", got, tc.want)
		}
	}
}

var users = []database.NewUser{
	{
		Email:                 "test@test.com",
		Username:              "test",
		DisplayName:           "test",
		Password:              "password",
		AvatarURL:             "avatar.jpg",
		EmailVerificationCode: "51235",
		EmailIsVerified:       false,
		FailIfNotInitialUser:  false,
		EnforcePasswordLength: false,
		TosAccepted:           false,
	},
	{
		Email:                 "test2@test.com",
		Username:              "test2",
		DisplayName:           "test2",
		Password:              "password",
		AvatarURL:             "avatar.jpg",
		EmailVerificationCode: "51235",
		EmailIsVerified:       false,
		FailIfNotInitialUser:  false,
		EnforcePasswordLength: false,
		TosAccepted:           false,
	},
	{
		Email:                 "test3@test.com",
		Username:              "test3",
		DisplayName:           "test3",
		Password:              "password",
		AvatarURL:             "avatar.jpg",
		EmailVerificationCode: "51235",
		EmailIsVerified:       false,
		FailIfNotInitialUser:  false,
		EnforcePasswordLength: false,
		TosAccepted:           false,
	},
	{
		Email:                 "test4@test.com",
		Username:              "test4",
		DisplayName:           "test4",
		Password:              "password",
		AvatarURL:             "avatar.jpg",
		EmailVerificationCode: "51235",
		EmailIsVerified:       false,
		FailIfNotInitialUser:  false,
		EnforcePasswordLength: false,
		TosAccepted:           false,
	},
}

var licensesToCreate = []struct {
	licenseId   string
	version     int
	licenseInfo license.Info
}{
	{
		licenseId: "b40537b3-d056-4235-afc2-1811cf9fa76e",
		version:   5,
		licenseInfo: license.Info{
			Tags:      []string{},
			UserCount: 10,
			ExpiresAt: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		licenseId: "9bbb0f96-b6c4-4545-926b-dd62ed3a7899",
		version:   12,
		licenseInfo: license.Info{
			Tags:      []string{},
			UserCount: 5,
			ExpiresAt: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		licenseId: "a8b0e28a-fc13-4724-b40a-12321202428b",
		version:   8,
		licenseInfo: license.Info{
			Tags:      []string{},
			UserCount: 30,
			ExpiresAt: time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		licenseId: "35e282b8-33c0-4eda-8225-8903a80e194f",
		version:   1,
		licenseInfo: license.Info{
			Tags:      []string{},
			UserCount: 27,
			ExpiresAt: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	},
}
