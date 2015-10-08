package notif

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/mattbaird/gochimp"
	"github.com/sourcegraph/go-ses"
)

const notifAdmin = "all@sourcegraph.com"
const notifFrom = "notify@sourcegraph.com"

var AwsEmailEnabled bool
var mandrillEnabled bool

var mandrill *gochimp.MandrillAPI

func init() {
	if mandrillKey := os.Getenv("MANDRILL_KEY"); mandrillKey != "" {
		mandrillEnabled = true

		var err error
		mandrill, err = gochimp.NewMandrill(mandrillKey)
		if err != nil {
			log.Panicf("could not initialize mandrill client: %s", err)
		}
	}
	AwsEmailEnabled, _ = strconv.ParseBool(os.Getenv("SG_SEND_NOTIFS"))
}

// SendAdminEmail sends email to Sourcegraph. It is for internal purposes
// only. For sending email to users, see 'SendMandrillTemplate'.
func SendAdminEmail(title string, body string) error {
	desc := fmt.Sprintf("From: %s\nTo: %s\nTitle: %s\nBody: %s\n", notifFrom, notifAdmin, title, body)
	if !AwsEmailEnabled {
		return fmt.Errorf("skipped sending email because SG_SEND_NOTIFS is false:\n%s", desc)
	}
	c := &ses.EnvConfig
	_, err := c.SendEmail(notifFrom, notifAdmin, title, body)
	if err != nil {
		return fmt.Errorf("error sending email notification: %s\n%s", err, desc)
	}
	log.Println("Email sent:", desc)
	return nil
}

// SendMandrillTemplate sends an email template through mandrill.
func SendMandrillTemplate(template, name, email string, mergeVars []gochimp.Var) {
	if !mandrillEnabled {
		log15.Info("skipped sending email because MANDRILL_KEY is empty", "template", template, "name", name, "email", email)
		return
	}
	go func() {
		responses, err := SendMandrillTemplateBlocking(template, name, email, mergeVars)
		if err != nil {
			log15.Error("Failed to send email through Mandrill", "template", template, "name", name, "email", email)
		} else if len(responses) != 1 {
			log15.Error("Unexpected responses from Mandrill", "template", template, "name", name, "email", email, "responses", responses)
		} else if responses[0].RejectedReason != "" {
			log15.Error("Email rejected by Mandrill", "template", template, "name", name, "email", email, "response", responses[0])
		}
	}()
}

// SendMandrillTemplateBlocking sends an email template through mandrill, but
// blocks until we have a response from Mandrill
func SendMandrillTemplateBlocking(template string, name string, email string, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
	if !mandrillEnabled {
		return nil, fmt.Errorf("skipped sending email because MANDRILL_KEY is empty:\nname: %s, email: %s", name, email)
	}
	return mandrill.MessageSendTemplate(template, nil, gochimp.Message{
		To:          []gochimp.Recipient{{Email: email, Name: name}},
		MergeVars:   []gochimp.MergeVars{{Recipient: email, Vars: mergeVars}},
		TrackOpens:  true,
		TrackClicks: true,
	}, false)
}
