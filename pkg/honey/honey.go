// package honey is a lightweight wrapper around libhoney which initializes
// honeycomb based on environment variables.
package honey

import (
	"log"
	"os"

	"github.com/sourcegraph/sourcegraph/pkg/env"

	libhoney "github.com/honeycombio/libhoney-go"
)

var writeKey = env.Get("HONEYCOMB_TEAM", "", "The key used for Honeycomb event tracking.")

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

// Builder creates a builder for logging to a dataset.
func Builder(dataset string) *libhoney.Builder {
	b := libhoney.NewBuilder()
	b.Dataset = dataset
	return b
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_832(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
