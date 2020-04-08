package app

import "github.com/sourcegraph/sourcegraph/internal/txemail"

func init() {
	txemail.DisableSilently()
}
