package userlimitchecker

import (
	"context"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

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
