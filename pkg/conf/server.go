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

// DefaultServerFrontendOnly is a server that should only ever be used by frontend.
// TODO@ggilmore: Write better description
var DefaultServerFrontendOnly *server

// server
type server struct {
	configFilePath string

	store *configStore

	needRestartMu sync.RWMutex
	needRestart   bool

	// fileWrite signals when our app writes to the configuration file. The
	// secondary channel is closed when server.Raw() would return the new
	// configuration that has been written to disk.
	// TODO@ggilmore: is it important that the channel here is buffered?
	// var fileWrite = make(chan chan struct{}, 1)
	fileWrite chan chan struct{}

	// ready is a barrier to block request handling until the server
	// has been initialized via server.start().
	ready chan struct{}

	once sync.Once
}

// Raw returns the raw text of the configuration file.
func (s *server) Raw() string {
	<-s.ready

	return s.store.Raw()
}

// TODO@ggilmore: Investigate if this is needed later.
func (s *server) IsDirty() bool {
	return false
}

// Write writes the JSON config file to the config file's path. If the JSON configuration is
// invalid, an error is returned.
func (s *server) Write(input string) error {
	<-s.ready

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
	// we would return to the caller earlier than server.Raw() would return the
	// new configuration.
	doneReading := make(chan struct{}, 1)
	s.fileWrite <- doneReading
	<-doneReading

	return nil
}

// Edit invokes the provided function to compute edits to the site
// configuration. It then applies and writes them.
//
// The computation function is provided the current configuration, which should
// NEVER be modified in any way. Always copy values.
func (s *server) Edit(computeEdits func(current *schema.SiteConfiguration, raw string) ([]jsonx.Edit, error)) error {
	<-s.ready

	// TODO@ggilmore: There is a race condition here (also present in the existing library).
	// Current and raw could be inconsistent. Another thing to offload to configStore?
	// Snapshot method?
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

	// TODO@ggilmore: Another race condition (also present in the existing library). Locks
	// aren't held between applying the edits and writing the config file,
	// so the newConfig could be outdated.
	err = s.Write(newConfig)
	if err != nil {
		return errors.Wrap(err, "conf.Write")
	}

	return nil
}

// start prepares the server to start handling requests by
// periodically reloading the configuration file from disk.
func (s *server) Start() {
	s.once.Do(func() {
		for {
			// TODO@ggilmore: This logic is incorrect. If there is a JSON syntax error when parsing the file,
			// then the channel will never be closed. (We block writing invalid files to disk with server.Write(), but
			// it's possible for people to directly edit the file). Maybe check to see if err is specifically
			// a syntax error, and continue anyway?
			//
			// Actualy, I am not sure if this is really a regression? We'd fail completly in the old version if
			// the site configuration was invalid. Maybe we should do that here too?
			err := s.updateFromDisk(false)
			if err == nil {
				close(s.ready)
				break
			}

			log.Printf("received error during initial configuration update, err: %s", err)
			time.Sleep(1 * time.Second)
		}

		go s.watchDisk()
	})
}

// watchDisk reloads the configuration file from disk at least every five seconds or whenever
// server.Write() is called.
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

	if !configChange.Changed {
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
// (bypasses the configStore).
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
