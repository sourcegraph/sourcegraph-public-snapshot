package graphqlbackend

import (
	"context"
	"net/url"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func makeUsersCreateTestDB() (*database.MockDB, *database.MockAuthzStore) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	users.CreateFunc.SetDefaultReturn(&types.User{ID: 1, Username: "alice"}, nil)

	authz := database.NewMockAuthzStore()
	authz.GrantPendingPermissionsFunc.SetDefaultReturn(nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	db.AuthzFunc.SetDefaultReturn(authz)
	return db, authz
}

func TestCreateUser(t *testing.T) {
	db, authz := makeUsersCreateTestDB()

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Query: `
				mutation {
					createUser(username: "alice") {
						user {
							id
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"createUser": {
						"user": {
							"id": "VXNlcjox"
						}
					}
				}
			`,
		},
	})

	mockrequire.Called(t, authz.GrantPendingPermissionsFunc)
}

func TestCreateUserResetPasswordURL(t *testing.T) {
	backend.MockMakePasswordResetURL = func(_ context.Context, _ int32) (*url.URL, error) {
		return url.Parse("/reset-url?code=foobar")
	}
	userpasswd.MockResetPasswordEnabled = func() bool { return true }
	t.Cleanup(func() {
		backend.MockMakePasswordResetURL = nil
		userpasswd.MockResetPasswordEnabled = nil
	})

	t.Run("with SMTP disabled", func(t *testing.T) {
		db, authz := makeUsersCreateTestDB()

		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				EmailSmtp: nil,
			},
		})
		t.Cleanup(func() { conf.Mock(nil) })

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation {
						createUser(username: "alice",email:"alice@sourcegraph.com",verifiedEmail:false) {
							user {
								id
							}
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: `
					{
						"createUser": {
							"user": {
								"id": "VXNlcjox"
							},
							"resetPasswordURL": "http://example.com/reset-url?code=foobar"
						}
					}
				`,
			},
		})

		mockrequire.Called(t, authz.GrantPendingPermissionsFunc)
	})

	t.Run("with SMTP enabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				EmailSmtp: &schema.SMTPServerConfig{},
			},
		})

		var sentMessage txemail.Message
		txemail.MockSend = func(_ context.Context, message txemail.Message) error {
			sentMessage = message
			return nil
		}
		t.Cleanup(func() {
			conf.Mock(nil)
			txemail.MockSend = nil
		})

		db, authz := makeUsersCreateTestDB()
		userEmails := database.NewMockUserEmailsStore()
		db.UserEmailsFunc.SetDefaultReturn(userEmails)

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation {
						createUser(username: "alice",email:"alice@sourcegraph.com",verifiedEmail:false) {
							user {
								id
							}
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: `
					{
						"createUser": {
							"user": {
								"id": "VXNlcjox"
							},
							"resetPasswordURL": "http://example.com/reset-url?code=foobar"
						}
					}
				`,
			},
		})

		data := sentMessage.Data.(userpasswd.SetPasswordEmailTemplateData)
		assert.Contains(t, data.URL, "http://example.com/reset-url")
		assert.Contains(t, data.URL, "&emailVerifyCode=")
		assert.Contains(t, data.URL, "&email=")

		mockrequire.Called(t, authz.GrantPendingPermissionsFunc)
		mockrequire.Called(t, userEmails.SetLastVerificationFunc)
	})

	t.Run("with SMTP enabled, without verifiedEmail", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				EmailSmtp: &schema.SMTPServerConfig{},
			},
		})

		var sentMessage txemail.Message
		txemail.MockSend = func(_ context.Context, message txemail.Message) error {
			sentMessage = message
			return nil
		}
		t.Cleanup(func() {
			conf.Mock(nil)
			txemail.MockSend = nil
		})

		db, authz := makeUsersCreateTestDB()
		userEmails := database.NewMockUserEmailsStore()
		db.UserEmailsFunc.SetDefaultReturn(userEmails)

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation {
						createUser(username: "alice",email:"alice@sourcegraph.com") {
							user {
								id
							}
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: `
					{
						"createUser": {
							"user": {
								"id": "VXNlcjox"
							},
							"resetPasswordURL": "http://example.com/reset-url?code=foobar"
						}
					}
				`,
			},
		})

		// should not have tried to issue email verification
		data := sentMessage.Data.(userpasswd.SetPasswordEmailTemplateData)
		assert.Contains(t, data.URL, "http://example.com/reset-url")
		assert.NotContains(t, data.URL, "&emailVerifyCode=")
		assert.NotContains(t, data.URL, "&email=")

		mockrequire.Called(t, authz.GrantPendingPermissionsFunc)
		mockrequire.NotCalled(t, userEmails.SetLastVerificationFunc)
	})
}
