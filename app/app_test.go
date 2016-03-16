package app

import "sourcegraph.com/sourcegraph/sourcegraph/notif"

func init() {
	notif.MustBeDisabled()
	Init()
}
