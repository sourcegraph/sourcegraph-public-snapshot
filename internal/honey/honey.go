// package honey is a lightweight wrapper around libhoney which initializes
// honeycomb based on environment variables.
package honey

import (
	"log"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"

	"github.com/honeycombio/libhoney-go"
)

var (
	apiKey  = env.Get("HONEYCOMB_TEAM", "", "The key used for Honeycomb event tracking.")
	suffix  = env.Get("HONEYCOMB_SUFFIX", "", "Suffix to append to honeycomb datasets. Used to differentiate between prod/dogfood/dev/etc.")
	disable = env.Get("HONEYCOMB_DISABLE", "", "Ignore that HONEYCOMB_TEAM is set and return false for Enabled. Used by specific instrumentation which ignores what Enabled returns and will log based on other criteria.")
)

// Enabled returns true if honeycomb has been configured to run.
func Enabled() bool {
	return apiKey != "" && disable == ""
}

func init() {
	if apiKey == "" {
		return
	}
	err := libhoney.Init(libhoney.Config{
		APIKey: apiKey,
	})
	if err != nil {
		log.Println("Failed to init libhoney:", err)
		apiKey = ""
		return
	}
	// HOSTNAME is the name of the pod on kubernetes.
	if h := hostname.Get(); h != "" {
		libhoney.AddField("pod_name", h)
	}
}
