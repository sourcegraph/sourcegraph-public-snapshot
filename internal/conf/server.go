package conf

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

// ConfigurationSource provides direct access to read and write to the
// "raw" configuration.
type ConfigurationSource interface {
	// Write updates the configuration. The Deployment field is ignored.
	Write(ctx context.Context, data conftypes.RawUnified) error
	Read(ctx context.Context) (conftypes.RawUnified, error)
}

// Server provides access and manages modifications to the site configuration.
type Server struct {
	Source ConfigurationSource

	store *store

	needRestartMu sync.RWMutex
	needRestart   bool

	// fileWrite signals when our app writes to the configuration file. The
	// secondary channel is closed when server.Raw() would return the new
	// configuration that has been written to disk.
	fileWrite chan chan struct{}

	once sync.Once
}

// NewServer returns a new Server instance that mangages the site config file
// that is stored at configSource.
//
// The server must be started with Start() before it can handle requests.
func NewServer(source ConfigurationSource) *Server {
	fileWrite := make(chan chan struct{}, 1)
	return &Server{
		Source:    source,
		store:     newStore(),
		fileWrite: fileWrite,
	}
}

// Raw returns the raw text of the configuration file.
func (s *Server) Raw() conftypes.RawUnified {
	return s.store.Raw()
}

// Write writes the JSON config file to the config file's path. If the JSON configuration is
// invalid, an error is returned.
func (s *Server) Write(ctx context.Context, input conftypes.RawUnified) error {
	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	_, err := ParseConfig(input)
	if err != nil {
		return err
	}

	err = s.Source.Write(ctx, input)
	if err != nil {
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

// Edits describes some JSON edits to apply to site configuration.
type Edits struct {
	Site []jsonx.Edit
}

// Edit invokes the provided function to compute edits to the site
// configuration. It then applies and writes them.
//
// The computation function is provided the current configuration, which should
// NEVER be modified in any way. Always copy values.
//
// TODO(slimsag): Currently, edits may only be applied via the frontend. It may
// make sense to allow non-frontend services to apply edits as well. To do this
// we would need to pipe writes through the frontend's internal httpapi.
func (s *Server) Edit(ctx context.Context, computeEdits func(current *Unified, raw conftypes.RawUnified) (Edits, error)) error {
	// TODO@ggilmore: There is a race condition here (also present in the existing library).
	// Current and raw could be inconsistent. Another thing to offload to configStore?
	// Snapshot method?
	current := s.store.LastValid()
	raw := s.store.Raw()

	// Compute edits.
	edits, err := computeEdits(current, raw)
	if err != nil {
		return errors.Wrap(err, "computeEdits")
	}

	// Apply edits and write out new configuration.
	newSite, err := jsonx.ApplyEdits(raw.Site, edits.Site...)
	if err != nil {
		return errors.Wrap(err, "jsonx.ApplyEdits Site")
	}

	// TODO@ggilmore: Another race condition (also present in the existing library). Locks
	// aren't held between applying the edits and writing the config file,
	// so the newConfig could be outdated.
	err = s.Write(ctx, conftypes.RawUnified{Site: newSite})
	if err != nil {
		return errors.Wrap(err, "conf.Write")
	}
	return nil
}

// Start initializes the server instance.
func (s *Server) Start() {
	s.once.Do(func() {
		go s.watchSource()
	})
}

// watchSource reloads the configuration from the source at least every five seconds or whenever
// server.Write() is called.
func (s *Server) watchSource() {
	ctx := context.Background()
	for {
		err := s.updateFromSource(ctx)
		if err != nil {
			log.Printf("failed to read configuration: %s. Fix your Sourcegraph configuration to resolve this error. Visit https://docs.sourcegraph.com/ to learn more.", err)
		}

		jitter := time.Duration(rand.Int63n(5 * int64(time.Second)))

		var signalDoneReading chan struct{}
		select {
		case signalDoneReading = <-s.fileWrite:
			// File was changed on FS, so check now.
		case <-time.After(jitter):
			// File possibly changed on FS, so check now.
		}

		if signalDoneReading != nil {
			close(signalDoneReading)
		}
	}
}

func (s *Server) updateFromSource(ctx context.Context) error {
	rawConfig, err := s.Source.Read(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to read configuration")
	}

	configChange, err := s.store.MaybeUpdate(rawConfig)
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
	if NeedRestartToApply(configChange.Old, configChange.New) {
		s.markNeedServerRestart()
	}

	return nil
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
