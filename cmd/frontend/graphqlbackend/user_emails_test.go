package graphqlbackend

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go/errors"
	"github.com/graph-gophers/graphql-go/gqltesting"

	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSetUserEmailVerified(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{SiteAdmin: true}, nil
	}
	db.Mocks.UserEmails.SetVerified = func(context.Context, int32, string, bool) error {
		return nil
	}

	tests := []struct {
		name                                string
		gqlTests                            []*gqltesting.Test
		expectCalledGrantPendingPermissions bool
	}{
		{
			name: "set an email to be verified",
			gqlTests: []*gqltesting.Test{
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
			gqlTests: []*gqltesting.Test{
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
			db.Mocks.Authz.GrantPendingPermissions = func(context.Context, *db.GrantPendingPermissionsArgs) error {
				calledGrantPendingPermissions = true
				return nil
			}

			gqltesting.RunTests(t, test.gqlTests)

			if test.expectCalledGrantPendingPermissions != calledGrantPendingPermissions {
				t.Fatalf("calledGrantPendingPermissions: want %v but got %v", test.expectCalledGrantPendingPermissions, calledGrantPendingPermissions)
			}
		})
	}
}

func TestResendUserEmailVerification(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByID = func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{ID: id, SiteAdmin: true}, nil
	}
	db.Mocks.Users.GetByCurrentAuthUser = func(context.Context) (*types.User, error) {
		return &types.User{ID: 1, SiteAdmin: true}, nil
	}
	db.Mocks.UserEmails.SetLastVerification = func(context.Context, int32, string, string) error {
		return nil
	}

	knownTime := time.Time{}.Add(1337 * time.Hour)
	timeNow = func() time.Time {
		return knownTime
	}

	tests := []struct {
		name            string
		gqlTests        []*gqltesting.Test
		email           *db.UserEmail
		expectEmailSent bool
	}{
		{
			name: "resend a verification email",
			gqlTests: []*gqltesting.Test{
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
			email: &db.UserEmail{
				Email:  "alice@example.com",
				UserID: 1,
			},
			expectEmailSent: true,
		},
		{
			name: "email already verified",
			gqlTests: []*gqltesting.Test{
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
			email: &db.UserEmail{
				Email:      "alice@example.com",
				UserID:     1,
				VerifiedAt: &knownTime,
			},
			expectEmailSent: false,
		},
		{
			name: "invalid email",
			gqlTests: []*gqltesting.Test{
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
					ExpectedErrors: []*errors.QueryError{
						{
							Message:       "oh no!",
							Path:          []interface{}{"resendVerificationEmail"},
							ResolverError: fmt.Errorf("oh no!"),
						},
					},
				},
			},
			email: &db.UserEmail{
				Email:      "alice@example.com",
				UserID:     1,
				VerifiedAt: &knownTime,
			},
			expectEmailSent: false,
		},
		{
			name: "resend a verification email, too soon",
			gqlTests: []*gqltesting.Test{
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
					ExpectedErrors: []*errors.QueryError{
						{
							Message:       "Last email sent too recently",
							Path:          []interface{}{"resendVerificationEmail"},
							ResolverError: fmt.Errorf("Last email sent too recently"),
						},
					},
				},
			},
			email: &db.UserEmail{
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
			db.Mocks.UserEmails.Get = func(id int32, email string) (string, bool, error) {
				if email != test.email.Email {
					return "", false, fmt.Errorf("oh no!")
				}
				return test.email.Email, test.email.VerifiedAt != nil, nil
			}
			db.Mocks.UserEmails.GetLatestVerificationSentEmail = func(context.Context, string) (*db.UserEmail, error) {
				return test.email, nil
			}

			gqltesting.RunTests(t, test.gqlTests)

			if emailSent != test.expectEmailSent {
				t.Errorf("Expected emailSent == %t, got %t", test.expectEmailSent, emailSent)
			}
		})
	}
}
