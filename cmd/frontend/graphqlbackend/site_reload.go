package graphqlbackend

import (
	"context"
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/processrestart"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// canReloadSite is whether the current site can be reloaded via the API. Currently
// only goreman-managed sites can be reloaded. Callers must also check if the actor
// is an admin before actually reloading the site.
var canReloadSite = processrestart.CanRestart()

func (r *schemaResolver) ReloadSite(ctx context.Context) (*EmptyResponse, error) {
	// 🚨 SECURITY: Reloading the site is an interruptive action, so only admins
	// may do it.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	if !canReloadSite {
		return nil, errors.New("reloading site is not supported")
	}

	const delay = 750 * time.Millisecond
	log15.Warn("Will reload site (from API request)", "actor", actor.FromContext(ctx))
	time.AfterFunc(delay, func() {
		log15.Warn("Reloading site", "actor", actor.FromContext(ctx))
		if err := processrestart.Restart(); err != nil {
			log15.Error("Error reloading site", "err", err)
		}
	})

	return &EmptyResponse{}, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_221(size int) error {
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
