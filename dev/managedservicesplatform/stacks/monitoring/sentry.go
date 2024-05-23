package monitoring

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/sentry/datasentryorganizationintegration"
	slackconversation "github.com/sourcegraph/managed-services-platform-cdktf/gen/slack/conversation"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/sentryalert"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func createSentryAlerts(
	stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	channel slackconversation.Conversation,
	slackIntegration datasentryorganizationintegration.DataSentryOrganizationIntegration,
) error {
	for _, config := range []sentryalert.Config{
		{
			ID:            "all-issues",
			SentryProject: vars.SentryProject,
			AlertConfig: sentryalert.AlertConfig{
				Name:        "Notify in Slack",
				Frequency:   60,                        // Notify for an issue at most once every hour,
				Conditions:  []sentryalert.Condition{}, // Empty conditions, always alert
				ActionMatch: sentryalert.ActionMatchAny,
				Actions: []sentryalert.Action{
					{
						ID: sentryalert.SlackNotifyServiceAction,
						ActionParameters: map[string]any{
							"workspace":  slackIntegration.Id(),
							"channel":    channel.Name(),
							"channel_id": channel.Id(),
							"tags": pointers.Stringf("msp-%s-%s",
								vars.Service.ID, vars.EnvironmentID),
						},
					},
				},
			},
		},
	} {
		if _, err := sentryalert.New(stack, id.Group(config.ID), config); err != nil {
			return errors.Wrapf(err, "creating sentry alert `%s`", config.ID)
		}
	}
	return nil
}
