package buildkite

const slackStepNotifyPluginName = "https://github.com/sourcegraph/step-slack-notify-buildkite-plugin.git#main"

// SlackStepNotifyConfigPayload represents the configuration for the SlackStepNotify plugin.
// For details over its configuration, see https://github.com/sourcegraph/step-slack-notify-buildkite-plugin
//
// TODO @jhchabran make that importable from the source, as those are written in Go.
type SlackStepNotifyConfigPayload struct {
	Message              string                           `json:"message"`
	ChannelName          string                           `json:"channel_name"`
	SlackTokenEnvVarName string                           `json:"slack_token_env_var_name"`
	Conditions           SlackStepNotifyPayloadConditions `json:"conditions"`
}

// SlackStepNotifyPayloadConditions represents a set of conditions that the plugin uses to evaluate
// if a notification should be sent or not.
//
// For more details, see https://github.com/sourcegraph/step-slack-notify-buildkite-plugin
type SlackStepNotifyPayloadConditions struct {
	ExitCodes []int    `json:"exit_codes,omitempty"`
	Failed    bool     `json:"failed,omitempty"`
	Branches  []string `json:"branches,omitempty"`
}

// SlackStepNotify enables to send a custom notification that depends only on a given step output, regardless
// of the final build output.
//
// Useful when used in conjuction with SoftFail, to keep running a flaky step and alerting its owners without
// disrupting the CI.
func SlackStepNotify(config *SlackStepNotifyConfigPayload) StepOpt {
	if config.SlackTokenEnvVarName == "" {
		// If no slack token is given, use the default which has the right permissions.
		config.SlackTokenEnvVarName = "CI_CUSTOM_SLACK_BUILDKITE_PLUGIN_TOKEN"
	}
	return flattenStepOpts(
		Plugin(slackStepNotifyPluginName, config),
	)
}
