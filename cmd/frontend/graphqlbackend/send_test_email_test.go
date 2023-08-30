package graphqlbackend

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	gqlerrors "github.com/graph-gophers/graphql-go/errors"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

const testUUID = "01234567-89ab-cdef-0123-456789abcdef"

func TestSendTestEmail(t *testing.T) {
	siteConfig := schema.SiteConfiguration{
		EmailSmtp: &schema.SMTPServerConfig{
			Host:           "smtp.test",
			Port:           587,
			Authentication: "PLAIN",
			Username:       "test",
			Password:       "password",
		},
		EmailAddress:    "test@example.com",
		EmailSenderName: "Test Example",
	}

	db := dbmocks.NewMockDB()
	users := dbmocks.NewStrictMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
	db.UsersFunc.SetDefaultReturn(users)

	s := mustParseGraphQLSchema(t, db)

	conf.Mock(&conf.Unified{SiteConfiguration: siteConfig})

	randomUUID = func() (uuid.UUID, error) {
		return uuid.MustParse(testUUID), nil
	}
	messageID := testUUID[:5]

	t.Cleanup(func() {
		conf.Mock(nil)
		txemail.MockSendWithConfig = nil
		randomUUID = uuid.NewRandom
	})
	tests := []struct {
		name      string
		to        string
		sendError *error
		config    *smtpConfig
		gqlTest   *Test
	}{{
		name: "Uses global config if no config provided in args",
		to:   "test@example.com",
		gqlTest: &Test{
			Schema: s,
			Query: `
				mutation {
					sendTestEmail(to: "test@example.com")
				}
				`,
			ExpectedResult: fmt.Sprintf(`
				{
					"sendTestEmail": "Sent test email to \"test@example.com\" successfully! Please check that it was received successfully. Compare the test ID on received email with: %s"
				}
				`, messageID),
		},
	}, {
		name: "Uses global config if no null config provided in args",
		to:   "test@example.com",
		gqlTest: &Test{
			Schema: s,
			Query: `
					mutation {
						sendTestEmail(to: "test@example.com", config: null)
					}
					`,
			ExpectedResult: fmt.Sprintf(`
					{
						"sendTestEmail": "Sent test email to \"test@example.com\" successfully! Please check that it was received successfully. Compare the test ID on received email with: %s"
					}
					`, messageID),
		},
	}, {
		name: "Uses conifg provided in args",
		to:   "test@example.com",
		config: &smtpConfig{
			Host:           "smtp.custom",
			Port:           587,
			Authentication: "none",
			EmailAddress:   "sourcegraph@custom.test",
		},
		gqlTest: &Test{
			Schema: s,
			Query: `
						mutation {
							sendTestEmail(to: "test@example.com", config: {
								host: "smtp.custom",
								port: 587,
								authentication: NONE,
								emailAddress: "sourcegraph@custom.test"
							})
						}
						`,
			ExpectedResult: fmt.Sprintf(`
						{
							"sendTestEmail": "Sent test email to \"test@example.com\" successfully! Please check that it was received successfully. Compare the test ID on received email with: %s"
						}
						`, messageID),
		},
	}, {
		name: "Uses conifg provided in args, with plain auth",
		to:   "test@example.com",
		config: &smtpConfig{
			Host:            "smtp.custom",
			Port:            587,
			Authentication:  "PLAIN",
			Username:        pointers.Ptr("alice"),
			Password:        pointers.Ptr("supersecret"),
			EmailAddress:    "sourcegraph@custom.test",
			EmailSenderName: pointers.Ptr("Sourcegraph Test"),
			NoVerifyTLS:     pointers.Ptr(true),
		},
		gqlTest: &Test{
			Schema: s,
			Query: `
							mutation {
								sendTestEmail(to: "test@example.com", config: {
									host: "smtp.custom",
									port: 587,
									authentication: PLAIN,
									username: "alice",
									password: "supersecret",
									emailAddress: "sourcegraph@custom.test",
									emailSenderName: "Sourcegraph Test",
									noVerifyTLS: true
								})
							}
							`,
			ExpectedResult: fmt.Sprintf(`
							{
								"sendTestEmail": "Sent test email to \"test@example.com\" successfully! Please check that it was received successfully. Compare the test ID on received email with: %s"
							}
							`, messageID),
		},
	}, {
		name: "Uses conifg provided in args and replaces REDACTED with saved values",
		to:   "test@example.com",
		config: &smtpConfig{
			Host:            "smtp.custom",
			Port:            587,
			Authentication:  "PLAIN",
			Username:        pointers.Ptr("test"),
			Password:        pointers.Ptr("password"),
			EmailAddress:    "sourcegraph@custom.test",
			EmailSenderName: pointers.Ptr("Sourcegraph Test"),
			NoVerifyTLS:     pointers.Ptr(true),
		},
		gqlTest: &Test{
			Schema: s,
			Query: `
								mutation {
									sendTestEmail(to: "test@example.com", config: {
										host: "smtp.custom",
										port: 587,
										authentication: PLAIN,
										username: "REDACTED",
										password: "REDACTED",
										emailAddress: "sourcegraph@custom.test",
										emailSenderName: "Sourcegraph Test",
										noVerifyTLS: true
									})
								}
								`,
			ExpectedResult: fmt.Sprintf(`
								{
									"sendTestEmail": "Sent test email to \"test@example.com\" successfully! Please check that it was received successfully. Compare the test ID on received email with: %s"
								}
								`, messageID),
		},
	}, {
		name:      "Returns error if email cannot be sent",
		to:        "test@example.com",
		sendError: pointers.Ptr(errors.New("error sending email")),
		gqlTest: &Test{
			Schema: s,
			Query: `
									mutation {
										sendTestEmail(to: "test@example.com")
									}
									`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Message: fmt.Sprintf("Failed to send test email: error sending email, look for test ID: %s", messageID),
					Path:    []interface{}{"sendTestEmail"},
				},
			},
		},
	},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{SiteConfiguration: siteConfig})
			var usedConfig schema.SiteConfiguration
			var to string

			txemail.MockSendWithConfig = func(_ context.Context, msg txtypes.Message, cfg schema.SiteConfiguration) error {
				to = msg.To[0]
				usedConfig = cfg
				err := pointers.Deref(test.sendError, nil)
				return err
			}

			t.Cleanup(func() {
				conf.Mock(nil)
				txemail.MockSendWithConfig = nil
			})

			RunTest(t, test.gqlTest)

			require.Equal(t, test.to, to)

			if test.config != nil {
				require.Equal(t, test.config.EmailAddress, usedConfig.EmailAddress)
				require.Equal(t, pointers.Deref(test.config.EmailSenderName, ""), usedConfig.EmailSenderName)
				require.Equal(t, test.config.Host, usedConfig.EmailSmtp.Host)
				require.Equal(t, int(test.config.Port), usedConfig.EmailSmtp.Port)
				require.Equal(t, pointers.Deref(test.config.Username, ""), usedConfig.EmailSmtp.Username)
				require.Equal(t, pointers.Deref(test.config.Password, ""), usedConfig.EmailSmtp.Password)
				require.Equal(t, strings.ToLower(string(test.config.Authentication)), strings.ToLower(usedConfig.EmailSmtp.Authentication))
				require.Equal(t, pointers.Deref(test.config.NoVerifyTLS, false), usedConfig.EmailSmtp.NoVerifyTLS)
			}
		})
	}
}
