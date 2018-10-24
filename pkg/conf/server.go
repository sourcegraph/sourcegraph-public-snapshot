package conf

import (
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/schema"
)

type server struct {
	configFilePath string

	cache *configCache

	needRestartMu sync.RWMutex
	needRestart   bool

	fileWrite chan chan struct{}
}

// fileWrite signals when our app writes to the configuration file. The
// secondary channel is closed when conf.Get() would return the new
// configuration that has been written to disk.
// TODO@ggilmore: is it important that the channel here is buffered?
// var fileWrite = make(chan chan struct{}, 1)

func (s *server) Raw() string {
	return s.cache.Raw()
}

func (s *server) Write(input string) error {
	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	after, err := parseConfig(input)
	if err != nil {
		return err
	}

	before := s.cache.Parsed()

	if err := ioutil.WriteFile(s.configFilePath, []byte(input), 0600); err != nil {
		return err
	}

	// Wait for the change to the configuration file to be detected. Otherwise
	// we would return to the caller earlier than conf.Get() would return the
	// new configuration.
	doneReading := make(chan struct{}, 1)
	s.fileWrite <- doneReading
	<-doneReading

	// Update global "needs restart" state.
	if needRestartToApply(before, after) {
		s.markNeedServerRestart()
	}

	return nil
}

// Edit invokes the provided function to compute edits to the site
// configuration. It then applies and writes them.
//
// The computation function is provided the current configuration, which should
// NEVER be modified in any way. Always copy values.
func (s *server) Edit(computeEdits func(current *schema.SiteConfiguration, raw string) ([]jsonx.Edit, error)) error {
	current := s.cache.Parsed()
	raw := s.cache.Raw()

	// Compute edits.
	edits, err := computeEdits(current, raw)
	if err != nil {
		return errors.Wrap(err, "computeEdits")
	}

	// Apply edits and write out new configuration.
	newConfig, err := jsonx.ApplyEdits(raw, edits...)
	if err != nil {
		return errors.Wrap(err, "jsonx.ApplyEdits")
	}

	err = s.Write(newConfig)
	if err != nil {
		return errors.Wrap(err, "conf.Write")
	}

	return nil
}

func (s *server) watchDisk() {
	for {
		var signalDoneReading chan struct{}
		select {
		case signalDoneReading = <-s.fileWrite:
			// File was changed on FS, so check now.
		case <-time.After(5 * time.Second):
			// File possibly changed on FS, so check now.
		}

		if IsDirty() {
			// Read the new configuration from disk.
			if err := initConfig(true); err != nil {
				log.Printf("failed to read configuration from environment: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://about.sourcegraph.com/docs to learn more.", err, configFilePath)
			}
		}

		if signalDoneReading != nil {
			close(signalDoneReading)
		}
	}
}

// NeedServerRestart tells if the server needs to restart for pending configuration
// changes to take effect.
func (s *server) NeedServerRestart() bool {
	s.needRestartMu.RLock()
	defer s.needRestartMu.RUnlock()
	return s.needRestart
}

// markNeedServerRestart marks the server as needing a restart so that pending
// configuration changes can take effect.
func (s *server) markNeedServerRestart() {
	s.needRestartMu.Lock()
	s.needRestart = true
	s.needRestartMu.Unlock()
}
