package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSetUserEmailVerified(t *testing.T) {
	resetMocks()
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	database.Mocks.UserEmails.SetVerified = func(context.Context, int32, string, bool) error {
		return nil
	}

	tests := []struct {
		name                                string
		gqlTests                            []*Test
		expectCalledGrantPendingPermissions bool
	}{
		{
			name: "set an email to be verified",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				mutation {
					setUserEmailVerified(user: "VXNlcjox", email: "alice@example.com", verified: true) {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"setUserEmailVerified": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			expectCalledGrantPendingPermissions: true,
		},
		{
			name: "set an email to be unverified",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				mutation {
					setUserEmailVerified(user: "VXNlcjox", email: "alice@example.com", verified: false) {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"setUserEmailVerified": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			expectCalledGrantPendingPermissions: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calledGrantPendingPermissions := false
			database.Mocks.Authz.GrantPendingPermissions = func(context.Context, *database.GrantPendingPermissionsArgs) error {
				calledGrantPendingPermissions = true
				return nil
			}

			RunTests(t, test.gqlTests)

			if test.expectCalledGrantPendingPermissions != calledGrantPendingPermissions {
				t.Fatalf("calledGrantPendingPermissions: want %v but got %v", test.expectCalledGrantPendingPermissions, calledGrantPendingPermissions)
			}
		})
	}
}

func TestResendUserEmailVerification(t *testing.T) {
	resetMocks()
	database.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, SiteAdmin: true}, nil
	}
	database.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{ID: 1, SiteAdmin: true}, nil
	}
	database.Mocks.UserEmails.SetLastVerification = func(context.Context, int32, string, string) error {
		return nil
	}

	knownTime := time.Time{}.Add(1337 * time.Hour)
	timeNow = func() time.Time {
		return knownTime
	}

	tests := []struct {
		name            string
		gqlTests        []*Test
		email           *database.UserEmail
		expectEmailSent bool
	}{
		{
			name: "resend a verification email",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alice@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"resendVerificationEmail": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			email: &database.UserEmail{
				Email:  "alice@example.com",
				UserID: 1,
			},
			expectEmailSent: true,
		},
		{
			name: "email already verified",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alice@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: `
				{
					"resendVerificationEmail": {
						"alwaysNil": null
    				}
				}
			`,
				},
			},
			email: &database.UserEmail{
				Email:      "alice@example.com",
				UserID:     1,
				VerifiedAt: &knownTime,
			},
			expectEmailSent: false,
		},
		{
			name: "invalid email",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alan@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: "null",
					ExpectedErrors: []*gqlerrors.QueryError{
						{
							Path:          []interface{}{"resendVerificationEmail"},
							Message:       "oh no!",
							ResolverError: errors.New("oh no!"),
						},
					},
				},
			},
			email: &database.UserEmail{
				Email:      "alice@example.com",
				UserID:     1,
				VerifiedAt: &knownTime,
			},
			expectEmailSent: false,
		},
		{
			name: "resend a verification email, too soon",
			gqlTests: []*Test{
				{
					Schema: mustParseGraphQLSchema(t),
					Query: `
				mutation {
					resendVerificationEmail(user: "VXNlcjox", email: "alice@example.com") {
						alwaysNil
					}
				}
			`,
					ExpectedResult: "null",
					ExpectedErrors: []*gqlerrors.QueryError{
						{
							Message:       "Last verification email sent too recently",
							Path:          []interface{}{"resendVerificationEmail"},
							ResolverError: errors.New("Last verification email sent too recently"),
						},
					},
				},
			},
			email: &database.UserEmail{
				Email:  "alice@example.com",
				UserID: 1,
				LastVerificationSentAt: func() *time.Time {
					t := knownTime.Add(-30 * time.Second)
					return &t
				}(),
			},
			expectEmailSent: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var emailSent bool
			txemail.MockSend = func(ctx context.Context, msg txemail.Message) error {
				emailSent = true
				return nil
			}
			database.Mocks.UserEmails.Get = func(id int32, email string) (string, bool, error) {
				if email != test.email.Email {
					return "", false, errors.New("oh no!")
				}
				return test.email.Email, test.email.VerifiedAt != nil, nil
			}
			database.Mocks.UserEmails.GetLatestVerificationSentEmail = func(context.Context, string) (*database.UserEmail, error) {
				return test.email, nil
			}

			RunTests(t, test.gqlTests)

			if emailSent != test.expectEmailSent {
				t.Errorf("Expected emailSent == %t, got %t", test.expectEmailSent, emailSent)
			}
		})
	}
}
