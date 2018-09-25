package userpasswd

import (
	"reflect"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Watch for configuration changes related to the builtin auth provider.
func init() {
	var (
		init = true

		mu sync.Mutex
		pc *schema.BuiltinAuthProvider
		pi auth.Provider
	)
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// Only react when the config changes.
		newPC, _ := getProviderConfig()
		if reflect.DeepEqual(newPC, pc) {
			return
		}

		if !init {
			log15.Info("Reloading changed builtin authentication provider configuration.")
		}
		updates := map[auth.Provider]bool{}
		var newPI auth.Provider
		if newPC != nil {
			newPI = &provider{c: newPC}
			updates[newPI] = true
		}
		if pi != nil {
			updates[pi] = false
		}
		auth.UpdateProviders(updates)
		pc = newPC
		pi = newPI
	})
	init = false
}
