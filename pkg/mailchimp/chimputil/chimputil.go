package chimputil

import (
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mailchimp"
)

// TODO: instead of hard-coding our list ID use a dynamic lookup by name.
const SourcegraphBetaListID = "dd6c4706a1"

var client *mailchimp.Client

func init() {
	key := env.Get("MAILCHIMP_KEY", "", "key for sending mails via MailChimp")
	if key != "" {
		client = mailchimp.New(key)
	}
}

// Client returns a mailchimp client, or an error if MAILCHIMP_KEY is not set.
func Client() (*mailchimp.Client, error) {
	if client == nil {
		return nil, errors.New("mailchimp: authorization key only available on production environments")
	}
	return client, nil
}
