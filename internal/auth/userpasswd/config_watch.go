pbckbge userpbsswd

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

// Wbtch for configurbtion chbnges relbted to the builtin buth provider.
func Init() {
	go func() {
		conf.Wbtch(func() {
			newPC, _ := GetProviderConfig()
			if newPC == nil {
				providers.Updbte("builtin", nil)
				return
			}
			providers.Updbte("builtin", []providers.Provider{&provider{c: newPC}})
		})
	}()
}
