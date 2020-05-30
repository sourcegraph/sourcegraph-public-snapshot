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

	needServerRestartMu sync.RWMutex
	needServerRestart   bool

	// fileWrite signals when our app writes to the configuration file. The
	// secondary channel is closed when server.Raw() would return the new
	// configuration that has been written to disk.
	fileWrite chan chan PostConfigWriteActions

	once sync.Once
}

// NewServer returns a new Server instance that mangages the site config file
// that is stored at configSource.
//
// The server must be started with Start() before it can handle requests.
func NewServer(source ConfigurationSource) *Server {
	fileWrite := make(chan chan PostConfigWriteActions, 1)
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
func (s *Server) Write(ctx context.Context, input conftypes.RawUnified) (PostConfigWriteActions, error) {
	actions := PostConfigWriteActions{}

	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	_, err := ParseConfig(input)
	if err != nil {
		return actions, err
	}

	err = s.Source.Write(ctx, input)
	if err != nil {
		return actions, err
	}

	// Wait for the change to the configuration file to be detected. Otherwise
	// we would return to the caller earlier than server.Raw() would return the
	// new configuration.
	doneReading := make(chan PostConfigWriteActions, 1)
	s.fileWrite <- doneReading
	actions = <-doneReading

	return actions, nil
}

// Edits describes some JSON edits to apply to site or critical configuration.
type Edits struct {
	Site, Critical []jsonx.Edit
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
	newCritical, err := jsonx.ApplyEdits(raw.Critical, edits.Critical...)
	if err != nil {
		return errors.Wrap(err, "jsonx.ApplyEdits Critical")
	}
	newSite, err := jsonx.ApplyEdits(raw.Site, edits.Site...)
	if err != nil {
		return errors.Wrap(err, "jsonx.ApplyEdits Site")
	}

	// TODO@ggilmore: Another race condition (also present in the existing library). Locks
	// aren't held between applying the edits and writing the config file,
	// so the newConfig could be outdated.
	if _, err := s.Write(ctx, conftypes.RawUnified{
		Site:     newSite,
		Critical: newCritical,
	}); err != nil {
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
		jitter := time.Duration(rand.Int63n(5 * int64(time.Second)))

		var signalDoneReading chan PostConfigWriteActions
		select {
		case signalDoneReading = <-s.fileWrite:
			// File was changed on FS, so check now.
		case <-time.After(jitter):
			// File possibly changed on FS, so check now.
		}

		actions, err := s.updateFromSource(ctx)
		if err != nil {
			log.Printf("failed to read configuration: %s. Fix your Sourcegraph configuration to resolve this error. Visit https://docs.sourcegraph.com/ to learn more.", err)
		}

		if signalDoneReading != nil {
			signalDoneReading <- actions
			close(signalDoneReading)
		}
	}
}

func (s *Server) updateFromSource(ctx context.Context) (PostConfigWriteActions, error) {
	actions := PostConfigWriteActions{}

	rawConfig, err := s.Source.Read(ctx)
	if err != nil {
		return actions, errors.Wrap(err, "unable to read configuration")
	}

	configChange, err := s.store.MaybeUpdate(rawConfig)
	if err != nil {
		return actions, err
	}

	// Don't need to restart if the configuration hasn't changed.
	if !configChange.Changed {
		return actions, nil
	}

	// Don't restart if the configuration was empty before (this only occurs during initialization).
	if configChange.Old == nil {
		return actions, nil
	}

	// Update global "action has to be taken for the configuration to apply"
	// state.
	actions = NeedActionToApply(configChange.Old, configChange.New)
	if actions.ServerRestartRequired {
		s.markNeedServerRestart()
	}
	// We don't persist the frontend reload state here because we can't monitor
	// if the user has actually done it: instead, we'll just report that
	// synchronously now in the return value, but otherwise drop it.

	return actions, nil
}

// NeedServerRestart tells if the server needs to restart for pending configuration
// changes to take effect.
func (s *Server) NeedServerRestart() bool {
	s.needServerRestartMu.RLock()
	defer s.needServerRestartMu.RUnlock()
	return s.needServerRestart
}

// markNeedServerRestart marks the server as needing a restart so that pending
// configuration changes can take effect.
func (s *Server) markNeedServerRestart() {
	s.needServerRestartMu.Lock()
	defer s.needServerRestartMu.Unlock()
	s.needServerRestart = true
}
