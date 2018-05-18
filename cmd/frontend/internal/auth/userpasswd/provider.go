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
		pi *auth.Provider
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
		newPI := newProviderInstance(newPC)
		auth.UpdateProviders(map[*auth.Provider]bool{newPI: true, pi: false})
		pc = newPC
		pi = newPI
	})
	init = false
}

func newProviderInstance(pc *schema.BuiltinAuthProvider) *auth.Provider {
	if pc == nil {
		return nil
	}

	return &auth.Provider{
		ProviderID: auth.ProviderID{ServiceType: pc.Type},
		Public: auth.PublicProviderInfo{
			DisplayName: "Builtin username-password authentication",
			IsBuiltin:   true,
		},
	}
}
