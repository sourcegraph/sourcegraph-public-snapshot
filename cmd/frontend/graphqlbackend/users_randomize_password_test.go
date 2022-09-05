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
	userID := int32(42)
	userIDBase64 := string(MarshalUserID(userID))

	db := database.NewMockDB()
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
						Path:    []any{"randomizeUserPassword"},
					},
				},
				Variables: map[string]any{"user": userIDBase64},
			},
		})
	})

	t.Run("Errors on Cloud when sending emails is not enabled", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}})
		defer conf.Mock(nil)

		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

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
						Path:    []any{"randomizeUserPassword"},
					},
				},
				Variables: map[string]any{"user": userIDBase64},
			},
		})
	})

	// tests below depend on AuthProviders and EmailSmtp being configured properly
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}, EmailSmtp: &schema.SMTPServerConfig{}}})
	defer conf.Mock(nil)

	t.Run("Returns error if user is not site-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)
		db.UsersFunc.SetDefaultReturn(users)

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
						Path:    []any{string("randomizeUserPassword")},
					},
				},
				Variables: map[string]any{"user": userIDBase64},
			},
		})
	})

	t.Run("Returns error when cannot parse user ID", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
		db.UsersFunc.SetDefaultReturn(users)

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
						Path:    []any{string("randomizeUserPassword")},
					},
				},
				Variables: map[string]any{"user": "alice"},
			},
		})
	})

	t.Run("Returns resetPasswordUrl if user is site-admin", func(t *testing.T) {
		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
		users.RandomizePasswordAndClearPasswordResetRateLimitFunc.SetDefaultReturn(nil)
		users.RenewPasswordResetCodeFunc.SetDefaultReturn("code", nil)
		db.UsersFunc.SetDefaultReturn(users)

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
				Variables: map[string]any{"user": userIDBase64},
			},
		})
	})

	t.Run("Does not return resetPasswordUrl when in Cloud", func(t *testing.T) {
		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig)

		users := database.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
		users.RandomizePasswordAndClearPasswordResetRateLimitFunc.SetDefaultReturn(nil)
		users.RenewPasswordResetCodeFunc.SetDefaultReturn("code", nil)
		users.GetByIDFunc.SetDefaultReturn(&types.User{Username: "alice"}, nil)

		userEmails := database.NewMockUserEmailsStore()
		userEmails.GetPrimaryEmailFunc.SetDefaultReturn("alice@foo.bar", false, nil)

		db.UsersFunc.SetDefaultReturn(users)
		db.UserEmailsFunc.SetDefaultReturn(userEmails)

		txemail.MockSend = func(ctx context.Context, message txemail.Message) error {
			return nil
		}
		defer func() {
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
				Variables: map[string]any{"user": userIDBase64},
			},
		})
	})
}
