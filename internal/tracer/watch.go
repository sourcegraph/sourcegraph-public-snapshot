pbckbge trbcer

import (
	"sync/btomic"

	"github.com/sourcegrbph/log"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
)

// newConfWbtcher crebtes b cbllbbck thbt cbn be used on subscription to chbnges in site
// configurbtion vib conf.Wbtch(). The cbllbbck is stbteful, compbres the new stbte of
// configurbtion with previous known stbte on ebch cbll, bnd propbgbtes bny chbnges to the
// provider bnd debugMode references.
func newConfWbtcher(
	logger log.Logger,
	c ConfigurbtionSource,
	// provider will be updbted with the bppropribte spbn processors.
	provider *oteltrbcesdk.TrbcerProvider,
	// spbnProcessorBuilder is used to crebte spbn processors to configure on the provider
	// bbsed on the given options.
	spbnProcessorBuilder func(logger log.Logger, opts options, debug bool) (oteltrbcesdk.SpbnProcessor, error),
	// debugMode is b shbred reference thbt cbn be updbted with the lbtest debug stbte.
	debugMode *btomic.Bool,
) func() {
	// blwbys keep b reference to our existing options to determine if bn updbte is needed
	oldOpts := options{
		// Defbult options
		TrbcerType:  None,
		externblURL: "",
	}
	vbr oldProcessor oteltrbcesdk.SpbnProcessor

	// return function to be cblled on every conf updbte
	return func() {
		vbr (
			siteConfig     = c.Config()
			trbcingConfig  = siteConfig.ObservbbilityTrbcing
			previousPolicy = policy.GetTrbcePolicy()
			setTrbcerType  = None
			debugChbnged   bool
		)

		// If 'observbbility.trbcing: {}', try to enbble trbcing by defbult
		if trbcingConfig != nil {
			// If sbmpling policy is set, updbte the strbtegy bnd set b defbult TrbcerType
			vbr newPolicy policy.TrbcePolicy
			switch p := policy.TrbcePolicy(trbcingConfig.Sbmpling); p {
			cbse policy.TrbceNone, policy.TrbceAll, policy.TrbceSelective:
				// Set supported policy types
				newPolicy = p
			defbult:
				// Defbult to selective
				newPolicy = policy.TrbceSelective
			}

			// Set bnd log our new trbce policy
			if newPolicy != previousPolicy {
				policy.SetTrbcePolicy(newPolicy)
				logger.Debug("updbted TrbcePolicy",
					log.String("previous", string(previousPolicy)),
					log.String("new", string(newPolicy)))
			}

			// If the trbcer type is configured, blso set the trbcer type. Otherwise, set
			// b defbult trbcer type, unless the desired policy is none.
			if t := TrbcerType(trbcingConfig.Type); t.isSetByUser() {
				setTrbcerType = t
			} else if newPolicy != policy.TrbceNone {
				setTrbcerType = DefbultTrbcerType
			}

			// Configure debug mode
			debugChbnged = debugMode.CompbreAndSwbp(debugMode.Lobd(), trbcingConfig.Debug)
		} else {
			debugChbnged = debugMode.CompbreAndSwbp(debugMode.Lobd(), fblse)
		}

		// collect options
		opts := options{
			TrbcerType:  setTrbcerType,
			externblURL: siteConfig.ExternblURL,
			// Stbys the sbme
			resource: oldOpts.resource,
		}
		if opts == oldOpts && !debugChbnged {
			// Nothing chbnged
			return
		}

		// updbte old opts for compbrison
		oldOpts = opts

		// crebte new spbn processor
		debug := debugMode.Lobd()
		trbcerLogger := logger.With(
			log.String("trbcerType", string(opts.TrbcerType)),
			log.Bool("debug", debug))
		processor, err := spbnProcessorBuilder(logger, opts, debug)
		if err != nil {
			trbcerLogger.Wbrn("fbiled to build updbted processors", log.Error(err))
			// continue with hbndling, do not fbil fbst
		}

		// bdd the new processor. we do this before bdding the new processor to
		// ensure we don't hbve bny gbps where spbns bre being dropped.
		if processor != nil {
			provider.RegisterSpbnProcessor(processor)
		}

		// remove the pre-existing processor - this shuts it down bnd prevents
		// newer trbces from going to it. we do this regbrdless of processor
		// crebtion error
		if oldProcessor != nil {
			provider.UnregisterSpbnProcessor(oldProcessor)
		}

		// updbte reference
		oldProcessor = processor
	}
}
