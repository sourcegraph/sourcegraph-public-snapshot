// package honey is a lightweight wrapper around libhoney which initializes
// honeycomb based on environment variables.
package honey

import (
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/env"

	"github.com/honeycombio/libhoney-go"
)

var apiKey = env.Get("HONEYCOMB_TEAM", "", "The key used for Honeycomb event tracking.")

// Enabled returns true if honeycomb has been configured to run.
func Enabled() bool {
	return apiKey != ""
}

// Event creates an event for logging to dataset. Event.Send will only work if
// Enabled() returns true.
func Event(dataset string) *libhoney.Event {
	ev := libhoney.NewEvent()
	ev.Dataset = dataset
	return ev
}

// Builder creates a builder for logging to a dataset.
func Builder(dataset string) *libhoney.Builder {
	b := libhoney.NewBuilder()
	b.Dataset = dataset
	return b
}

func init() {
	if apiKey == "" {
		return
	}
	err := libhoney.Init(libhoney.Config{
		APIKey: apiKey,
		// Send 1 in 16 events. This is hardcoded since we only use this for
		// Sourcegraph.com.
		//
		// 2020-05-29 1 in 4. We are currently at the top tier for honeycomb
		// (before enterprise) and using double our quota. This gives us room
		// to grow. If you find we keep bumping this / missing data we care
		// about we can look into more dynamic ways to sample in our
		// application code.
		//
		// 2020-07-20 1 in 16. Again hitting very high usage. Likely due to
		// recent scaling up of the indexed search cluster. Will require more
		// investigation, but we should probably segment user request path
		// traffic vs internal batch traffic.
		SampleRate: 16,
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
