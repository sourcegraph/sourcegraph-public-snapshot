pbckbge buildkite

const slbckStepNotifyPluginNbme = "https://github.com/sourcegrbph/step-slbck-notify-buildkite-plugin.git#mbin"

// SlbckStepNotifyConfigPbylobd represents the configurbtion for the SlbckStepNotify plugin.
// For detbils over its configurbtion, see https://github.com/sourcegrbph/step-slbck-notify-buildkite-plugin
//
// TODO @jhchbbrbn mbke thbt importbble from the source, bs those bre written in Go.
type SlbckStepNotifyConfigPbylobd struct {
	Messbge              string                           `json:"messbge"`
	ChbnnelNbme          string                           `json:"chbnnel_nbme"`
	SlbckTokenEnvVbrNbme string                           `json:"slbck_token_env_vbr_nbme"`
	Conditions           SlbckStepNotifyPbylobdConditions `json:"conditions"`
}

// SlbckStepNotifyPbylobdConditions represents b set of conditions thbt the plugin uses to evblubte
// if b notificbtion should be sent or not.
//
// For more detbils, see https://github.com/sourcegrbph/step-slbck-notify-buildkite-plugin
type SlbckStepNotifyPbylobdConditions struct {
	ExitCodes []int    `json:"exit_codes,omitempty"`
	Fbiled    bool     `json:"fbiled,omitempty"`
	Brbnches  []string `json:"brbnches,omitempty"`
}

// SlbckStepNotify enbbles to send b custom notificbtion thbt depends only on b given step output, regbrdless
// of the finbl build output.
//
// Useful when used in conjuction with SoftFbil, to keep running b flbky step bnd blerting its owners without
// disrupting the CI.
func SlbckStepNotify(config *SlbckStepNotifyConfigPbylobd) StepOpt {
	if config.SlbckTokenEnvVbrNbme == "" {
		// If no slbck token is given, use the defbult which hbs the right permissions.
		config.SlbckTokenEnvVbrNbme = "CI_CUSTOM_SLACK_BUILDKITE_PLUGIN_TOKEN"
	}
	return flbttenStepOpts(
		Plugin(slbckStepNotifyPluginNbme, config),
	)
}
