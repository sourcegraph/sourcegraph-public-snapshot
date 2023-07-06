package userlimitchecker

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegraph/log"
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
