package conf

import (
	"context"
	"sync"

	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	source ConfigurationSource
	// sourceWrites signals when our app writes to the configuration source. The
	// received channel should be closed when server.Raw() would return the new
	// configuration that has been written to disk.
	sourceWrites chan chan struct{}

	needRestartMu sync.RWMutex
	needRestart   bool

	startOnce sync.Once
}

// NewServer returns a new Server instance that mangages the site config file
// that is stored at configSource.
//
// The server must be started with Start() before it can handle requests.
func NewServer(source ConfigurationSource) *Server {
	return &Server{
		source:       source,
		sourceWrites: make(chan chan struct{}, 1),
	}
}

// Write validates and writes input to the server's source.
func (s *Server) Write(ctx context.Context, input conftypes.RawUnified) error {
	// Parse the configuration so that we can diff it (this also validates it
	// is proper JSON).
	_, err := ParseConfig(input)
	if err != nil {
		return err
	}

	err = s.source.Write(ctx, input)
	if err != nil {
		return err
	}

	// Wait for the change to the configuration file to be detected. Otherwise
	// we would return to the caller earlier than server.Raw() would return the
	// new configuration.
	doneReading := make(chan struct{}, 1)
	// Notify that we've written an update
	s.sourceWrites <- doneReading
	// Get notified that the update has been read (it gets closed) - don't write
	// until this is done.
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
	client := DefaultClient()
	current := client.store.LastValid()
	raw := client.Raw()

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
	s.startOnce.Do(func() {
		// We prepare to watch for config updates in order to mark the config server as
		// needing a restart (or not). This must be in a goroutine, since Watch must
		// happen after conf initialization (which may have not happened yet)
		go func() {
			var oldConfig *Unified
			Watch(func() {
				// Don't indicate restarts if this is the first update (initial configuration
				// after service startup).
				if oldConfig == nil {
					oldConfig = Get()
					return
				}

				// Update global "needs restart" state.
				newConfig := Get()
				if needRestartToApply(oldConfig, newConfig) {
					s.markNeedServerRestart()
				}

				// Update old value
				oldConfig = newConfig
			})
		}()
	})
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
