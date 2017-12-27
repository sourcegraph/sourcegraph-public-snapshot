package notif

import (
	"errors"
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mattbaird/gochimp"
)

var disableSilently bool

var mandrill *gochimp.MandrillAPI

func init() {
	if conf.CanSendEmail() && conf.Get().MandrillKey != "" {
		var err error
		mandrill, err = gochimp.NewMandrill(conf.Get().MandrillKey)
		if err != nil {
			log.Panicf("could not initialize mandrill client: %s", err)
		}
	}
}

// DisableSilently prevents sending of emails, even if the env var is set.
// Use it in tests to ensure that they do not send live notifications.
func DisableSilently() {
	disableSilently = true
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
	if disableSilently {
		log15.Info("skipped sending email because Mandrill key is empty", "config", config)
		return
	}
	if mandrill == nil {
		log15.Error("failed to send email because Mandrill key is empty", "config", config)
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
	if disableSilently {
		return nil, fmt.Errorf("skipped sending email because Mandrill key is empty: %#v", config)
	}
	if mandrill == nil {
		return nil, errors.New("failed to send email because Mandrill key is empty")
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
