package userlimitchecker

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/sourcegraph/log"
	ps "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/dotcom/productsubscription"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendApproachingUserLimitAlert(t *testing.T) {
	logger := log.NoOp()
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	// create user to satisfy product_subscription foreign key constraint
	userStore := db.Users()
	user, err := userStore.Create(ctx, users[0])
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

	t.Run("sends correctly formatted email", func(t *testing.T) {
		var gotEmail txemail.Message
		txemail.MockSend = func(ctx context.Context, message txemail.Message) error {
			gotEmail = message
			return nil
		}
		t.Cleanup(func() { txemail.MockSend = nil })

		err = SendApproachingUserLimitAlert(ctx, db)
		require.NoError(t, err)

		replyTo := "support@sourcegraph.com"
		messageId := "approaching_user_limit"
		want := &txemail.Message{
			To:        []string{"test@test.com"},
			Template:  approachingUserLimitEmailTemplate,
			MessageID: &messageId,
			ReplyTo:   &replyTo,
			Data: SetApproachingUserLimitTemplateData{
				RemainingUsers: 1,
			},
		}

		assert.Equal(t, want.To, gotEmail.To)
		assert.Equal(t, approachingUserLimitEmailTemplate, gotEmail.Template)
		assert.Equal(t, want.MessageID, gotEmail.MessageID)
		assert.Equal(t, want.ReplyTo, gotEmail.ReplyTo)
		assert.Equal(t, want.MessageID, gotEmail.MessageID)
		gotEmailData := want.Data.(SetApproachingUserLimitTemplateData)
		assert.Equal(t, 1, gotEmailData.RemainingUsers)
	})

	t.Run("does not send email if user count is not approaching user limit", func(t *testing.T) {
		err := licensesStore.UpdateUserCount(ctx, id, "15")
		require.NoError(t, err)

		old := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		err = SendApproachingUserLimitAlert(ctx, db)
		if err != nil {
			t.Errorf("could not run sendApproachingUserLimitAlert function: %s", err)
		}

		w.Close()
		out, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		os.Stdout = old

		assert.Equal(t, "user count on license within limit\n", string(out))
	})

	t.Run("does not send email if email sent within 7 days of current time", func(t *testing.T) {
		err := licensesStore.UpdateUserCount(ctx, id, "2")
		require.NoError(t, err)

		err = licensesStore.UpdateUserCountAlertSentAt(ctx, id, time.Now().UTC())
		require.NoError(t, err)

		old := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		err = SendApproachingUserLimitAlert(ctx, db)
		require.NoError(t, err)

		w.Close()
		out, err := ioutil.ReadAll(r)
		require.NoError(t, err)
		os.Stdout = old

		assert.Equal(t, "email recently sent\n", string(out))
	})
}

func TestGetPercentOfLimit(t *testing.T) {
	cases := []struct {
		expected  int
		userCount int
		userLimit int
	}{
		{expected: 0, userCount: 0, userLimit: 100},
		{expected: 84, userCount: 211, userLimit: 250},
		{expected: 61, userCount: 348, userLimit: 567},
		{expected: 46, userCount: 583, userLimit: 1264},
		{expected: 110, userCount: 10, userLimit: 0},
		{expected: 112, userCount: 45, userLimit: 40},
		{expected: 95, userCount: 19, userLimit: 20},
		{expected: 96, userCount: 87, userLimit: 90},
		{expected: 95, userCount: 95, userLimit: 100},
		{expected: 95, userCount: 3350, userLimit: 3500},
	}

	for _, tc := range cases {
		t.Run("should calculate correct percentage", func(t *testing.T) {
			actual := getPercentOfLimit(tc.userCount, tc.userLimit)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestGetLicenseUserLimit(t *testing.T) {
	logger := log.NoOp()
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	// create user to satisfy product_subscription foreign key constraint
	userStore := db.Users()
	user, err := userStore.Create(ctx, users[0])
	require.NoError(t, err)

	// create product_subscription to satisfy product_license foreign key constraint
	subStore := ps.NewDbSubscription(db)
	subId, err := subStore.Create(ctx, user.ID, user.Username)
	require.NoError(t, err)

	licensesStore := ps.NewDbLicense(db)
	for _, license := range licensesToCreate {
		_, err = licensesStore.Create(
			ctx,
			subId,
			license.licenseId,
			license.version,
			license.licenseInfo,
		)
		require.NoError(t, err)
	}

	actual, err := getLicenseUserLimit(ctx, db)
	require.NoError(t, err)

	expected := 30
	assert.Equal(t, expected, actual)
}

func TestGetUserCount(t *testing.T) {
	logger := log.NoOp()
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	userStore := db.Users()

	// create users state in db
	for _, user := range users {
		_, err := userStore.Create(ctx, user)
		require.NoError(t, err)
	}

	actual, err := getUserCount(ctx, db)
	require.NoError(t, err)

	expected := 4
	assert.Equal(t, expected, actual)
}

func TestGetVerifiedSiteAdminEmails(t *testing.T) {
	logger := log.NoOp()
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	userStore := db.Users()

	// Create users state in db
	var createdUsers []*types.User
	for i, user := range users {
		newUser, err := userStore.Create(ctx, user)
		require.NoError(t, err)
		createdUsers = append(createdUsers, newUser)

		// first, second and third created users are site admins
		if i <= 2 {
			userStore.SetIsSiteAdmin(ctx, createdUsers[i].ID, true)
		}
	}

	t.Run("should return slice of verified emails only", func(t *testing.T) {
		expected, err := getVerifiedSiteAdminEmails(ctx, db)
		require.NoError(t, err)
		actual := []string{"test@test.com", "test3@test.com"}
		assert.Equal(t, expected, actual)
	})
}

func TestGetUserEmail(t *testing.T) {
	logger := log.NoOp()
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	userStore := db.Users()

	cases := []struct {
		name     string
		expected string
		user     database.NewUser
	}{
		{
			name:     "should return email test@test.com",
			expected: "test@test.com",
			user:     users[0],
		},
		{
			name:     "should return email test2@test.com",
			expected: "test2@test.com",
			user:     users[1],
		},
		{
			name:     "should return email test3@test.com",
			expected: "test3@test.com",
			user:     users[2],
		},
		{
			name:     "should return email test4@test.com",
			expected: "test4@test.com",
			user:     users[3],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newUser, err := userStore.Create(ctx, tc.user)
			require.NoError(t, err)

			actual, _, err := getUserEmail(ctx, db, newUser)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, actual)
		})
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
		EmailIsVerified:       true,
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
		EmailIsVerified:       true,
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
