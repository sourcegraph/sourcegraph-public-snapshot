package backend

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"
)

// MakeEmailVerificationCode returns a random string that can be used as an email verification code.
func MakeEmailVerificationCode() string {
	emailCodeBytes := make([]byte, 20)
	if _, err := rand.Read(emailCodeBytes); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(emailCodeBytes)
}

// SendUserEmailVerificationEmail sends an email to the user to verify the email address. The code
// is the verification code that the user must provide to verify their access to the email address.
func SendUserEmailVerificationEmail(ctx context.Context, email, code string) error {
	q := make(url.Values)
	q.Set("code", code)
	q.Set("email", email)
	verifyEmailPath, _ := router.Router().Get(router.VerifyEmail).URLPath()
	return txemail.Send(ctx, txemail.Message{
		To:       []string{email},
		Template: verifyEmailTemplates,
		Data: struct {
			Email string
			URL   string
		}{
			Email: email,
			URL: globals.AppURL.ResolveReference(&url.URL{
				Path:     verifyEmailPath.Path,
				RawQuery: q.Encode(),
			}).String(),
		},
	})
}

var (
	verifyEmailTemplates = txemail.MustValidate(txemail.Templates{
		Subject: `Verify your email on Sourcegraph`,
		Text: `
Verify your email address {{printf "%q" .Email}} on Sourcegraph by following this link:

  {{.URL}}
`,
		HTML: `
<p>Verify your email address {{printf "%q" .Email}} on Sourcegraph to finish signing up.</p>

<p><strong><a href="{{.URL}}">Verify email address</a></p>
`,
	})
)
