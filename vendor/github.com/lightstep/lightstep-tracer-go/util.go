package lightstep

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var (
	seededGUIDGen     *rand.Rand
	seededGUIDGenOnce sync.Once
	seededGUIDLock    sync.Mutex
)

func genSeededGUID() string {
	// Golang does not seed the rng for us. Make sure it happens.
	seededGUIDGenOnce.Do(func() {
		seededGUIDGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	})

	// The golang random generators are *not* intrinsically thread-safe.
	seededGUIDLock.Lock()
	defer seededGUIDLock.Unlock()
	return strconv.FormatUint(uint64(seededGUIDGen.Int63()), 36)
}

var logOneError sync.Once

// maybeLogError logs the first error it receives using the standard log
// package and may also log subsequent errors based on verboseFlag.
func (r *Recorder) maybeLogError(err error) {
	if r.verbose {
		log.Printf("LightStep error: %v\n", err)
	} else {
		// Even if the flag is not set, always log at least one error.
		logOneError.Do(func() {
			log.Printf("LightStep instrumentation error (%v). Set the Verbose option to enable more logging.\n", err)
		})
	}
}

// maybeLogInfof may format and log its arguments if verboseFlag is set.
func (r *Recorder) maybeLogInfof(format string, args ...interface{}) {
	if r.verbose {
		s := fmt.Sprintf(format, args...)
		log.Printf("LightStep info: %s\n", s)
	}
}
