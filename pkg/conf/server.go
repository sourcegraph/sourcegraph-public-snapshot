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

	store *configStore

	needRestartMu sync.RWMutex
	needRestart   bool

	// fileWrite signals when our app writes to the configuration file. The
	// secondary channel is closed when conf.Get() would return the new
	// configuration that has been written to disk.
	// TODO@ggilmore: is it important that the channel here is buffered?
	// var fileWrite = make(chan chan struct{}, 1)
	fileWrite chan chan struct{}
}

func (s *server) Raw() string {
	return s.store.Raw()
}

func (s *server) Write(input string) error {
	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	_, err := parseConfig(input)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(s.configFilePath, []byte(input), 0600); err != nil {
		return err
	}

	// Wait for the change to the configuration file to be detected. Otherwise
	// we would return to the caller earlier than conf.Get() would return the
	// new configuration.
	doneReading := make(chan struct{}, 1)
	s.fileWrite <- doneReading
	<-doneReading

	// Update global "needs restart" state
	// TODO@ggilmore: Is this necessary? Why can't we rely on the background process to do this?.
	// if needRestartToApply(before, after) {
	// 	s.markNeedServerRestart()
	// }

	return nil
}

// Edit invokes the provided function to compute edits to the site
// configuration. It then applies and writes them.
//
// The computation function is provided the current configuration, which should
// NEVER be modified in any way. Always copy values.
func (s *server) Edit(computeEdits func(current *schema.SiteConfiguration, raw string) ([]jsonx.Edit, error)) error {
	current := s.store.Parsed()
	raw := s.store.Raw()

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

		err := s.updateFromDisk(true)
		if err != nil {
			log.Printf("failed to read configuration file: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://docs.sourcegraph.com/ to learn more.", err, s.configFilePath)
		}

		if signalDoneReading != nil {
			close(signalDoneReading)
		}
	}
}

func (s *server) updateFromDisk(reinitialize bool) error {
	rawConfig, err := s.readConfig()
	if err != nil {
		return err
	}

	configChange, err := s.store.MaybeUpdate(rawConfig)
	if err != nil {
		return err
	}

	if configChange == nil {
		return nil
	}

	if reinitialize {
		// Update global "needs restart" state.
		if needRestartToApply(configChange.Old, configChange.New) {
			s.markNeedServerRestart()
		}
	}

	return nil
}

// readConfig reads the raw configuration that's currently saved to the disk
// (bypasses the cache).
func (s *server) readConfig() (string, error) {
	data, err := ioutil.ReadFile(s.configFilePath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read config file from %q", s.configFilePath)
	}

	return string(data), nil
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

// FilePath is the path to the configuration file, if any.
// TODO@ggilmore: re-evaluate whether or not we need this
func (s *server) FilePath() string {
	return s.configFilePath
}
