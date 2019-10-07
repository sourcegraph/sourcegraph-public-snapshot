package app

import "github.com/sourcegraph/sourcegraph/cmd/internal/txemail"

func init() {
	txemail.DisableSilently()
}
