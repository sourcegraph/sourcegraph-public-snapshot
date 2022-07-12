package buildkite

const slackStepNotifyPluginName = "https://github.com/sourcegraph/step-slack-notify-buildkite-plugin.git#main"

// TODO @jhchabran make that importable from the source, as those are written in Go.
type SlackStepNotifyConfigPayload struct {
	Message              string                           `json:"message"`
	ChannelName          string                           `json:"channel_name"`
	SlackTokenEnvVarName string                           `json:"slack_token_env_var_name"`
	Conditions           SlackStepNotifyPayloadConditions `json:"conditions"`
}

type SlackStepNotifyPayloadConditions struct {
	ExitCodes []int    `json:"exit_codes,omitempty"`
	Failed    bool     `json:"failed,omitempty"`
	Branches  []string `json:"branches,omitempty"`
}

func SlackStepNotify(config *SlackStepNotifyConfigPayload) StepOpt {
	return flattenStepOpts(
		Plugin(slackStepNotifyPluginName, config),
	)
}
