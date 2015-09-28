package notif

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
func SendMandrillTemplate(template string, name string, email string, mergeVars []gochimp.Var) ([]gochimp.SendResponse, error) {
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
