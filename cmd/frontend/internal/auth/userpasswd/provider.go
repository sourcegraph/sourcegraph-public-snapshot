package userpasswd

import (
	"reflect"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	var (
		init = true

		mu sync.Mutex
		pc *schema.BuiltinAuthProvider
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
		pc = newPC
	})
	init = false
}
