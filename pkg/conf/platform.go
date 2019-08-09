package conf

import (
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/schema"
)

// PlatformConfiguration contains site configuration for the Sourcegraph platform.
type PlatformConfiguration struct {
	RemoteRegistryURL string
}

// DefaultRemoteRegistry is the default value for the site configuration property
// "extensions"."remoteRegistry".
//
// It is intentionally not set in the OSS build.
var DefaultRemoteRegistry string

// Extensions returns the configuration for the Sourcegraph platform, or nil if it is disabled.
func Extensions() *PlatformConfiguration {
	cfg := Get()

	x := cfg.Extensions
	if x == nil {
		if DefaultRemoteRegistry == "" {
			// There is no reasonable default behavior for extensions given that there is no remote
			// registry, so consider them disabled.
			return nil
		}

		x = &schema.Extensions{}
	}

	if x.Disabled != nil && *x.Disabled {
		return nil
	}

	var pc PlatformConfiguration

	// If the "remoteRegistry" value is a string, use that. If false, then keep it empty. Otherwise
	// use the default.
	if s, ok := x.RemoteRegistry.(string); ok {
		pc.RemoteRegistryURL = s
	} else if b, ok := x.RemoteRegistry.(bool); ok && !b {
		// Nothing to do.
	} else {
		pc.RemoteRegistryURL = DefaultRemoteRegistry
	}

	if v, _ := strconv.ParseBool(os.Getenv("OFFLINE")); v {
		pc.RemoteRegistryURL = ""
	}

	return &pc
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_727(size int) error {
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
