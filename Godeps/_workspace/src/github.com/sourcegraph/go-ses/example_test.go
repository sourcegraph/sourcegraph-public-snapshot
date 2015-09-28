package ses_test

import (
	"fmt"

	"github.com/sourcegraph/go-ses"
)

func Example() {
	// Change the From address to a sender address that is verified in your Amazon SES account.
	from := "notify@sourcegraph.com"
	to := "success@simulator.amazonses.com"

	// EnvConfig uses the AWS credentials in the environment variables $AWS_ACCESS_KEY_ID and
	// $AWS_SECRET_KEY.
	res, err := ses.EnvConfig.SendEmail(from, to, "Hello, world!", "Here is the message body.")
	if err == nil {
		fmt.Printf("Sent email: %s...\n", res[:32])
	} else {
		fmt.Printf("Error sending email: %s\n", err)
	}

	// output:
	// Sent email: <SendEmailResponse xmlns="http:/...
}
