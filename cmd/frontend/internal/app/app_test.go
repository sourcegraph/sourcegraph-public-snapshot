package app

import "github.com/sourcegraph/sourcegraph/pkg/txemail"

func init() {
	txemail.DisableSilently()
}
