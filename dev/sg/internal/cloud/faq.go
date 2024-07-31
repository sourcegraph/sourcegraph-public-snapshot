package cloud

import (
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
)

var faqEphemeralCommand = cli.Command{
	Name:        "faq",
	Usage:       "Opens the Cloud ephemeral FAQ",
	Description: "Opens the Cloud ephemeral FAQ",
	Action:      showFAQ,
}

func showFAQ(ctx *cli.Context) error {
	return open.URL(FAQLink)
}
