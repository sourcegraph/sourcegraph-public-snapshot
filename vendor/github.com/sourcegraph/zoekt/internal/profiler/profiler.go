package profiler

import (
	"log"
	"os"

	"cloud.google.com/go/profiler"
)

// Init starts the supported profilers IFF the environment variable is set.
func Init(svcName, version string, blockProfileRate int) {
	if os.Getenv("GOOGLE_CLOUD_PROFILER_ENABLED") != "" {
		err := profiler.Start(profiler.Config{
			Service:        svcName,
			ServiceVersion: version,
			MutexProfiling: true,
			AllocForceGC:   true,
		})
		if err != nil {
			log.Printf("could not initialize profiler: %s", err.Error())
		}
	}
}
