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

type EmailConfig struct {
	Template  string
	FromName  string
	FromEmail string
	ToName    string
	ToEmail   string
	Subject   string
}

// SendMandrillTemplate sends an email template through mandrill.
func SendMandrillTemplate(config *EmailConfig, templateContent []gochimp.Var, mergeVars []gochimp.Var) {
	if !mandrillEnabled {
		log15.Info("skipped sending email because MANDRILL_KEY is empty", "config", config)
		return
	}
	go func() {
		responses, err := SendMandrillTemplateBlocking(config, templateContent, mergeVars)
		if err != nil {
			log15.Error("Failed to send email through Mandrill", "config", config)
		} else if len(responses) != 1 {
			log15.Error("Unexpected responses from Mandrill", "config", config, "responses", responses)
		} else if responses[0].RejectedReason != "" {
			log15.Error("Email rejected by Mandrill", "config", config, "response", responses[0])
		}
	}()
}

// SendMandrillTemplateBlocking sends an email template through mandrill, but
// blocks until we have a response from Mandrill
func SendMandrillTemplateBlocking(config *EmailConfig, templateContent []gochimp.Var, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
	if !mandrillEnabled {
		return nil, fmt.Errorf("skipped sending email because MANDRILL_KEY is empty: %#v", config)
	}
	return mandrill.MessageSendTemplate(config.Template, templateContent, gochimp.Message{
		To:          []gochimp.Recipient{{Email: config.ToEmail, Name: config.ToName}},
		MergeVars:   []gochimp.MergeVars{{Recipient: config.ToEmail, Vars: mergeVars}},
		FromEmail:   config.FromEmail,
		FromName:    config.FromName,
		Subject:     config.Subject,
		TrackOpens:  true,
		TrackClicks: true,
	}, false)
}

// EmailIsConfigured returns true if the instance has an email configuration
func EmailIsConfigured() bool {
	return mandrillEnabled
}
