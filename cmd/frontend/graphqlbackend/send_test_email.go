package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

func (r *schemaResolver) SendTestEmail(ctx context.Context, args struct{ To string }) (string, error) {
	// 🚨 SECURITY: Only site admins can send test emails.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return "", err
	}

	logger := r.logger.Scoped("SendTestEmail")

	// Generate a simple identifier to make each email unique (don't need the full ID)
	var testID string
	if fullID, err := uuid.NewRandom(); err != nil {
		logger.Warn("failed to generate ID for test email", log.Error(err))
	} else {
		testID = fullID.String()[:5]
	}
	logger = logger.With(log.String("testID", testID))

	if err := txemail.Send(ctx, "test_email", txemail.Message{
		To:       []string{args.To},
		Template: emailTemplateTest,
		Data: struct {
			ID string
		}{
			ID: testID,
		},
	}); err != nil {
		logger.Error("failed to send test email", log.Error(err))
		return fmt.Sprintf("Failed to send test email: %s, look for test ID: %s", err, testID), nil
	}
	logger.Info("sent test email")

	return fmt.Sprintf("Sent test email to %q successfully! Please check it was received - look for test ID: %s",
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
