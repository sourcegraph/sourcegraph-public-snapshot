package conf

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
import "gopkg.in/inconshreveable/log15.v2"

var AppSecretKey = []byte(env.Get("SRC_APP_SECRET_KEY", "", "Private key used for symetric encryption of app payloads."))

func init() {
	if string(AppSecretKey) == "" {
		log15.Warn("SRC_APP_SECRET_KEY not set")
	}
}
