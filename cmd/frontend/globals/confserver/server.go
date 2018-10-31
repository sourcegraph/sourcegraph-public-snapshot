package confserver

import (
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf/parse"
	"github.com/sourcegraph/sourcegraph/pkg/conf/store"
)

// Server provides access and manages modifications to the site configuration.
type Server struct {
	// configFilePath is the path to the site configuration file on disk.
	basicFilePath string
	coreFilePath  string

	basicStore *store.BasicStore
	coreStore  *store.CoreStore

	// fileWriteBasic signals when our app writes to the configuration file. The
	// secondary channel is closed when server.RawBasic() would return the new
	// configuration that has been written to disk.
	fileWriteBasic chan chan struct{}
	// fileWriteCore signals when our app writes to the configuration file. The
	// secondary channel is closed when server.RawCore() would return the new
	// configuration that has been written to disk.
	fileWriteCore chan chan struct{}

	needRestartMu sync.RWMutex
	needRestart   bool

	once sync.Once
}

// NewServer returns a new Server instance that mangages the site config file
// that is stored at "configFilePath".
//
// The server must be started with Start() before it can handle requests.
func NewServer(basicFilePath, coreFilePath string) *Server {
	return &Server{
		basicFilePath:  basicFilePath,
		coreFilePath:   coreFilePath,
		basicStore:     store.NewBasicStore(),
		coreStore:      store.NewCoreStore(),
		fileWriteBasic: make(chan chan struct{}, 1),
		fileWriteCore:  make(chan chan struct{}, 1),
	}
}

// Raw returns the raw text of the configuration file.
func (s *Server) RawBasic() string {
	return s.basicStore.Raw()
}

// RawCor
func (s *Server) RawCore() string {
	return s.coreStore.Raw()
}

// Write writes the JSON config file to the config file's path. If the JSON configuration is
// invalid, an error is returned.
func (s *Server) WriteBasic(input string) error {
	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	_, err := parse.DeprecatedParseBasicConfigFromEnvironment(input)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(s.basicFilePath, []byte(input), 0600); err != nil {
		return err
	}

	// Wait for the change to the configuration file to be detected. Otherwise
	// we would return to the caller earlier than server.Raw() would return the
	// new configuration.
	doneReading := make(chan struct{}, 1)
	s.fileWriteBasic <- doneReading
	<-doneReading

	return nil
}

func (s *Server) WriteCore(input string) error {
	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	_, err := parse.DeprecatedParseCoreConfigFromEnvironment(input)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(s.coreFilePath, []byte(input), 0600); err != nil {
		return err
	}

	// Wait for the change to the configuration file to be detected. Otherwise
	// we would return to the caller earlier than server.Raw() would return the
	// new configuration.
	doneReading := make(chan struct{}, 1)
	s.fileWriteCore <- doneReading
	<-doneReading

	return nil
}

// Start initalizes the server instance.
func (s *Server) Start() {
	s.once.Do(func() {
		go s.watchDiskBasic()
		go s.watchDiskCore()
	})
}

// watchDiskBasic reloads the basic configuration file from disk at least
// every five seconds or whenever server.WriteBasic() is called.
func (s *Server) watchDiskBasic() {
	s.watchDisk(s.fileWriteBasic, s.basicFilePath, s.updateBasicFromDisk)
}

// watchDiskCore reloads the core configuration file from disk at least
// every five seconds or whenever server.WriteCore() is called.
func (s *Server) watchDiskCore() {
	s.watchDisk(s.fileWriteCore, s.coreFilePath, s.updateCoreFromDisk)
}

func (s *Server) watchDisk(fileWrite chan chan struct{}, configFilePath string, updateConfig func() error) {
	for {
		jitter := time.Duration(rand.Int63n(5 * int64(time.Second)))

		var signalDoneReading chan struct{}
		select {
		case signalDoneReading = <-fileWrite:
			// File was changed on FS, so check now.
		case <-time.After(jitter):
			// File possibly changed on FS, so check now.
		}

		err := updateConfig()
		if err != nil {
			log.Printf("failed to read configuration file: %s. Fix your Sourcegraph configuration (%s) to resolve this error. Visit https://docs.sourcegraph.com/ to learn more.", err, configFilePath)
		}

		if signalDoneReading != nil {
			close(signalDoneReading)
		}
	}
}

func (s *Server) updateCoreFromDisk() error {
	rawConfig, err := s.readConfig(s.coreFilePath)
	if err != nil {
		return err
	}

	configChange, err := s.coreStore.MaybeUpdate(rawConfig)
	if err != nil {
		return err
	}

	// Don't need to restart if the configuration hasn't changed.
	if !configChange.Changed {
		return nil
	}

	// Don't restart if the configuration was empty before (this only occurs during initialization).
	if configChange.Old == nil {
		return nil
	}

	// Update global "needs restart" state.
	if parse.NeedRestartToApplyCore(configChange.Old, configChange.New) {
		s.markNeedServerRestart()
	}

	return nil
}

func (s *Server) updateBasicFromDisk() error {
	rawConfig, err := s.readConfig(s.basicFilePath)
	if err != nil {
		return err
	}

	configChange, err := s.basicStore.MaybeUpdate(rawConfig)
	if err != nil {
		return err
	}

	// Don't need to restart if the configuration hasn't changed.
	if !configChange.Changed {
		return nil
	}

	// Don't restart if the configuration was empty before (this only occurs during initialization).
	if configChange.Old == nil {
		return nil
	}

	// Update global "needs restart" state.
	if parse.NeedRestartToApplyBasic(configChange.Old, configChange.New) {
		s.markNeedServerRestart()
	}

	return nil
}

// readConfig reads the raw configuration that's currently saved to the disk
// (bypasses the configStore).
func (s *Server) readConfig(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read config file from %q", path)
	}

	return string(data), nil
}

// NeedServerRestart tells if the server needs to restart for pending configuration
// changes to take effect.
func (s *Server) NeedServerRestart() bool {
	s.needRestartMu.RLock()
	defer s.needRestartMu.RUnlock()
	return s.needRestart
}

// markNeedServerRestart marks the server as needing a restart so that pending
// configuration changes can take effect.
func (s *Server) markNeedServerRestart() {
	s.needRestartMu.Lock()
	s.needRestart = true
	s.needRestartMu.Unlock()
}

// FilePath is the path to the configuration file, if any.
// TODO@ggilmore: re-evaluate whether or not we need this
func (s *Server) FilePathBasic() string {
	return s.basicFilePath
}

func (s *Server) FilePathCore() string {
	return s.coreFilePath
}
