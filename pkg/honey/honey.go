// package honey is a lightweight wrapper around libhoney which initializes
// honeycomb based on environment variables.
package honey

import (
	"log"
	"os"

	libhoney "github.com/honeycombio/libhoney-go"
)

var writeKey = os.Getenv("HONEYCOMB_TEAM")

// Enabled returns true if honeycomb has been configured to run.
func Enabled() bool {
	return writeKey != ""
}

// Event creates an event for logging to dataset. Event.Send will only work if
// Enabled() returns true.
func Event(dataset string) *libhoney.Event {
	ev := libhoney.NewEvent()
	ev.Dataset = dataset
	return ev
}

func init() {
	if writeKey == "" {
		return
	}
	err := libhoney.Init(libhoney.Config{
		WriteKey: writeKey,
	})
	if err != nil {
		log.Println("Failed to init libhoney:", err)
		writeKey = ""
		return
	}
	// HOSTNAME is the name of the pod on kubernetes.
	if h := os.Getenv("HOSTNAME"); h != "" {
		libhoney.AddField("pod_name", h)
	}
}
