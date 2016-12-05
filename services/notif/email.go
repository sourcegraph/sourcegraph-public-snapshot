package notif

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mattbaird/gochimp"
)

var mandrillEnabled bool

var mandrill *gochimp.MandrillAPI

var mandrillKey = env.Get("MANDRILL_KEY", "", "key for sending mails via Mandrill")

func init() {
	if mandrillKey != "" {
		mandrillEnabled = true

		var err error
		mandrill, err = gochimp.NewMandrill(mandrillKey)
		if err != nil {
			log.Panicf("could not initialize mandrill client: %s", err)
		}
	}
}

// Disable prevents sending of emails, even if the env var is set.
// Use it in tests to ensure that they do not send live notifications.
func Disable() {
	mandrillEnabled = false
}

// SendMandrillTemplate sends an email template through mandrill.
func SendMandrillTemplate(template, name, email, subject string, templateContent []gochimp.Var, mergeVars []gochimp.Var) {
	if !mandrillEnabled {
		log15.Info("skipped sending email because MANDRILL_KEY is empty", "template", template, "name", name, "email", email, "subject", subject)
		return
	}
	go func() {
		responses, err := SendMandrillTemplateBlocking(template, name, email, subject, templateContent, mergeVars)
		if err != nil {
			log15.Error("Failed to send email through Mandrill", "template", template, "name", name, "email", email, "subject", subject)
		} else if len(responses) != 1 {
			log15.Error("Unexpected responses from Mandrill", "template", template, "name", name, "email", email, "subject", subject, "responses", responses)
		} else if responses[0].RejectedReason != "" {
			log15.Error("Email rejected by Mandrill", "template", template, "name", name, "email", email, "subject", subject, "response", responses[0])
		}
	}()
}

// SendMandrillTemplateBlocking sends an email template through mandrill, but
// blocks until we have a response from Mandrill
func SendMandrillTemplateBlocking(template, name, email, subject string, templateContent []gochimp.Var, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
	if !mandrillEnabled {
		return nil, fmt.Errorf("skipped sending email because MANDRILL_KEY is empty:\nname: %s, email: %s", name, email)
	}
	return mandrill.MessageSendTemplate(template, templateContent, gochimp.Message{
		To:          []gochimp.Recipient{{Email: email, Name: name}},
		MergeVars:   []gochimp.MergeVars{{Recipient: email, Vars: mergeVars}},
		FromEmail:   "noreply@sourcegraph.com",
		FromName:    "Sourcegraph",
		Subject:     subject,
		TrackOpens:  true,
		TrackClicks: true,
	}, false)
}

// EmailIsConfigured returns true if the instance has an email configuration
func EmailIsConfigured() bool {
	return mandrillEnabled
}
