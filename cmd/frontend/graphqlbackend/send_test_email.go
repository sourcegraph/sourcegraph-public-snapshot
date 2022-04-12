package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

func (r *schemaResolver) SendTestEmail(ctx context.Context, args struct{ To string }) (string, error) {
	// 🚨 SECURITY: Only site admins can send test emails.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return "", err
	}

	if err := txemail.Send(ctx, txemail.Message{
		To:       []string{args.To},
		Template: emailTemplateTest,
		Data:     struct{}{},
	}); err != nil {
		return fmt.Sprintf("Failed to send test email: %s", err), nil
	}
	return fmt.Sprintf("Sent test email to %q successfully! Please check it was received.", args.To), nil
}

var emailTemplateTest = txemail.MustValidate(txtypes.Templates{
	Subject: `TEST: email sent from Sourcegraph`,
	Text: `
If you're seeing this, Sourcegraph is able to send email correctly for all of it's product features!

Congratulations!

* Sourcegraph
`,
	HTML: `
<p>Sourcegraph is able to send email correctly for all of it's product features!</p>
<br>
<p>Congratulations!</p>
<br>
<p>* Sourcegraph</p>
`,
})
