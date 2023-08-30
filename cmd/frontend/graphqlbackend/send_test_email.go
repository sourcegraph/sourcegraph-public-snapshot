package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

type SMTPAuthType string

type smtpConfig struct {
	Host            string       `json:"host"`
	Port            int32        `json:"port"`
	Authentication  SMTPAuthType `json:"authentication"`
	Username        *string      `json:"username,omitempty"`
	Password        *string      `json:"password,omitempty"`
	Domain          *string      `json:"domain,omitempty"`
	NoVerifyTLS     *bool        `json:"noVerifyTLS,omitempty"`
	EmailAddress    string       `json:"emailAddress"`
	EmailSenderName *string      `json:"emailSenderName,omitempty"`
}

var randomUUID = uuid.NewRandom

type SendTestEmailArgs struct {
	To     string
	Config *smtpConfig
}

func (r *schemaResolver) SendTestEmail(ctx context.Context, args SendTestEmailArgs) (string, error) {
	// 🚨 SECURITY: Only site admins can send test emails.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return "", err
	}

	logger := r.logger.Scoped("SendTestEmail", "email send test")

	// Generate a simple identifier to make each email unique (don't need the full ID)
	var testID string
	if fullID, err := randomUUID(); err != nil {
		logger.Warn("failed to generate ID for test email", log.Error(err))
	} else {
		testID = fullID.String()[:5]
	}
	logger = logger.With(log.String("testID", testID))

	var config schema.SiteConfiguration
	if args.Config != nil {
		// normalize authentication based on site.schema.json values
		authentication := string(args.Config.Authentication)
		if authentication == "CRAM_MD5" {
			authentication = "CRAM-MD5"
		} else if authentication == "NONE" {
			authentication = "none"
		}

		config = schema.SiteConfiguration{
			EmailSmtp: &schema.SMTPServerConfig{
				Host:           args.Config.Host,
				Port:           int(args.Config.Port),
				Authentication: authentication,
				Username:       pointers.Deref(args.Config.Username, ""),
				Password:       pointers.Deref(args.Config.Password, ""),
				Domain:         pointers.Deref(args.Config.Domain, ""),
				NoVerifyTLS:    pointers.Deref(args.Config.NoVerifyTLS, false),
			},
			EmailAddress:    args.Config.EmailAddress,
			EmailSenderName: pointers.Deref(args.Config.EmailSenderName, ""),
		}
		// replace redacted fields with actual values from site config
		if config.EmailSmtp.Username == "REDACTED" {
			config.EmailSmtp.Username = conf.SiteConfig().EmailSmtp.Username
		}
		if config.EmailSmtp.Password == "REDACTED" {
			config.EmailSmtp.Password = conf.SiteConfig().EmailSmtp.Password
		}
	} else {
		config = conf.SiteConfig()
	}

	if err := txemail.SendWithConfig(ctx, "test_email", txemail.Message{
		To:       []string{args.To},
		Template: emailTemplateTest,
		Data: struct {
			ID string
		}{
			ID: testID,
		},
	}, config); err != nil {
		logger.Error("failed to send test email", log.Error(err))
		return "", errors.Newf("Failed to send test email: %s, look for test ID: %s", err, testID)
	}
	logger.Info("sent test email")

	return fmt.Sprintf("Sent test email to %q successfully! Please check that it was received successfully. Compare the test ID on received email with: %s",
		args.To, testID), nil
}

var emailTemplateTest = txemail.MustValidate(txtypes.Templates{
	Subject: `TEST: email sent from Sourcegraph (test ID: {{ .ID }})`,
	Text: `
If you're seeing this, Sourcegraph is able to send email correctly for all of its product features!

Congratulations!

* Sourcegraph

Test ID: {{ .ID }}
`,
	HTML: `
<p>Sourcegraph is able to send email correctly for all of its product features!</p>
<br>
<p>Congratulations!</p>
<br>
<p>* Sourcegraph</p>
<br>
<p>Test ID: {{ .ID }}</p>
`,
})
