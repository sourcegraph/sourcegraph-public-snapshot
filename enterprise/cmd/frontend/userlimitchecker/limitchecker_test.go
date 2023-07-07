package userlimitchecker

import (
	"context"
	// "fmt"
	"reflect"
	"testing"
	// "time"

	"github.com/sourcegraph/log"
	// ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	// "github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
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

/* func TestGetUserLimit(t *testing.T) {
	ctx := context.Background()
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	licenses := ps.NewDbLicense(db)

	licensesToCreate := []struct {
		subscriptionId string
		licenseId      string
		version        int
		licenseInfo    license.Info
	}{
		{
			subscriptionId: "64440a21-acf5-4849-bd30-3910f9048316",
			licenseId:      "b40537b3-d056-4235-afc2-1811cf9fa76e",
			version:        5,
			licenseInfo: license.Info{
				Tags:      []string{},
				UserCount: 10,
				ExpiresAt: time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			subscriptionId: "64440a21-acf5-4849-bd30-3910f9048316",
			licenseId:      "b40537b3-d056-4235-afc2-1811cf9fa76e",
			version:        12,
			licenseInfo: license.Info{
				Tags:      []string{},
				UserCount: 5,
				ExpiresAt: time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			subscriptionId: "64440a21-acf5-4849-bd30-3910f9048316",
			licenseId:      "b40537b3-d056-4235-afc2-1811cf9fa76e",
			version:        8,
			licenseInfo: license.Info{
				Tags:      []string{},
				UserCount: 30,
				ExpiresAt: time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	var nl []string
	for _, license := range licensesToCreate {
		createdLicense, err := licenses.Create(
			ctx,
			license.subscriptionId,
			license.licenseId,
			license.version,
			license.licenseInfo,
		)
		if err != nil {
			t.Errorf("could not create new license: %s", err)
		}

		nl = append(nl, createdLicense)
	}

	for i, l := range nl {
		info, err := licenses.GetByID(ctx, l)
		if err != nil {
			t.Errorf("could not retrieve license info: %s", err)
		}

		if i == 1 {
			reason := fmt.Sprintf("test reason")
			info.RevokeReason = &reason
		}
	}

	got, err := getUserLimit(ctx, db)
	if err != nil {
		t.Errorf("could not get user limit: %s", err)
	}

	want := 5
	if got != want {
		t.Errorf("got %d want %d", got, want)
	}
} */

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
		if i == 0 {
			createdUsers[i].SiteAdmin = true
		}
	}

	got, _ := getSiteAdmins(ctx, db)
	want := []string{"test@test.com"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestGetUserEmail(t *testing.T) {
	logger := log.NoOp()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Background()
	userStore := db.Users()
	userInfo := database.NewUser{
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
	}

	newUser, err := userStore.Create(ctx, userInfo)
	if err != nil {
		t.Errorf("could not create new user: %s", err)
	}

	got, _, _ := getUserEmail(ctx, db, newUser)
	want := "test@test.com"
	if got != want {
		t.Errorf("got %s want %s", got, want)
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
