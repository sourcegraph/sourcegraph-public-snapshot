package app

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"

func init() {
	txemail.DisableSilently()
}
