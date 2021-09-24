// package honey is a lightweight wrapper around libhoney which initializes
// honeycomb based on environment variables.
package honey

import (
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/env"

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

// Event creates an event for logging to dataset. Event.Send will only work if
// Enabled() returns true.
func Event(dataset string) *libhoney.Event {
	ev := libhoney.NewEvent()
	ev.Dataset = dataset + suffix
	return ev
}

// EventWithFields creates an event for logging to the given dataset. The given
// fields are assigned to the event.
func EventWithFields(dataset string, fields map[string]interface{}) *libhoney.Event {
	ev := Event(dataset)
	for key, value := range fields {
		ev.AddField(key, value)
	}

	return ev
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
	if h := os.Getenv("HOSTNAME"); h != "" {
		libhoney.AddField("pod_name", h)
	}
}
