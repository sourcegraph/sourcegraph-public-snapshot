package app

import "sourcegraph.com/pkg/txemail"

func init() {
	txemail.DisableSilently()
}
