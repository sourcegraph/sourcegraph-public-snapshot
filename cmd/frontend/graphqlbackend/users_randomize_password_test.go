package graphqlbackend

import (
	"context"
	"testing"

	"github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestRandomizeUserPassword(t *testing.T) {
	resetMocks()

	defer func() {
		resetMocks()
	}()

	db := database.NewDB(nil)
	userID := int32(42)
	userIDBase64 := string(MarshalUserID(userID))

	t.Run("Errors when resetting passwords is not enabled", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation($user: ID!) {
						randomizeUserPassword(user: $user) {
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: "null",
				ExpectedErrors: []*errors.QueryError{
					{
						Message: "resetting passwords is not enabled",
						Path:    []interface{}{string("randomizeUserPassword")},
					},
				},
				Variables: map[string]interface{}{"user": userIDBase64},
			},
		})
	})

	t.Run("Errors on Cloud when sending emails is not enabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}})
		envvar.MockSourcegraphDotComMode(true)

		defer func() {
			conf.Mock(nil)
			envvar.MockSourcegraphDotComMode(false)
		}()

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation($user: ID!) {
						randomizeUserPassword(user: $user) {
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: "null",
				ExpectedErrors: []*errors.QueryError{
					{
						Message: "unable to reset password because email sending is not configured",
						Path:    []interface{}{string("randomizeUserPassword")},
					},
				},
				Variables: map[string]interface{}{"user": userIDBase64},
			},
		})
	})

	// tests below depend on AuthProviders and EmailSmtp being configured properly
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}, EmailSmtp: &schema.SMTPServerConfig{}}})

	t.Run("Returns error if user is not site-admin", func(t *testing.T) {
		db := database.NewDB(nil)
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: false}, nil
		}
		defer resetMocks()

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation($user: ID!) {
						randomizeUserPassword(user: $user) {
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: "null",
				ExpectedErrors: []*errors.QueryError{
					{
						Message: "must be site admin",
						Path:    []interface{}{string("randomizeUserPassword")},
					},
				},
				Variables: map[string]interface{}{"user": userIDBase64},
			},
		})
	})

	t.Run("Returns error when cannot parse user ID", func(t *testing.T) {
		db := database.NewDB(nil)
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: true}, nil
		}

		defer resetMocks()

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation($user: ID!) {
						randomizeUserPassword(user: $user) {
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: "null",
				ExpectedErrors: []*errors.QueryError{
					{
						Message: "cannot parse user ID: illegal base64 data at input byte 4",
						Path:    []interface{}{string("randomizeUserPassword")},
					},
				},
				Variables: map[string]interface{}{"user": "alice"},
			},
		})
	})

	t.Run("Returns resetPasswordUrl if user is site-admin", func(t *testing.T) {
		db := database.NewDB(nil)
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: true}, nil
		}
		database.Mocks.Users.RandomizePasswordAndClearPasswordResetRateLimit = func(ctx context.Context, userID int32) error {
			return nil
		}
		database.Mocks.Users.RenewPasswordResetCode = func(ctx context.Context, id int32) (string, error) {
			return "code", nil
		}

		defer resetMocks()

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation($user: ID!) {
						randomizeUserPassword(user: $user) {
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: `{
					"randomizeUserPassword": {
						"resetPasswordURL": "http://example.com/password-reset?code=code&userID=42"
					}
				}`,
				Variables: map[string]interface{}{"user": userIDBase64},
			},
		})
	})

	t.Run("Does not return resetPasswordUrl when in Cloud", func(t *testing.T) {
		envvar.MockSourcegraphDotComMode(true)

		db := database.NewDB(nil)
		database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
			return &types.User{SiteAdmin: true}, nil
		}
		database.Mocks.Users.RandomizePasswordAndClearPasswordResetRateLimit = func(ctx context.Context, userID int32) error {
			return nil
		}
		database.Mocks.Users.RenewPasswordResetCode = func(ctx context.Context, id int32) (string, error) {
			return "code", nil
		}
		database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
			return &types.User{
				Username: "alice",
			}, nil
		}
		database.Mocks.UserEmails.GetPrimaryEmail = func(ctx context.Context, id int32) (email string, verified bool, err error) {
			return "alice@foo.bar", false, nil
		}

		txemail.MockSend = func(ctx context.Context, message txemail.Message) error {
			return nil
		}

		defer func() {
			resetMocks()
			envvar.MockSourcegraphDotComMode(false)
			txemail.MockSend = nil
		}()

		RunTests(t, []*Test{
			{
				Schema: mustParseGraphQLSchema(t, db),
				Query: `
					mutation($user: ID!) {
						randomizeUserPassword(user: $user) {
							resetPasswordURL
						}
					}
				`,
				ExpectedResult: `{
					"randomizeUserPassword": {
						"resetPasswordURL": null
					}
				}`,
				Variables: map[string]interface{}{"user": userIDBase64},
			},
		})
	})
}
