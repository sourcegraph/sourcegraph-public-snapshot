package app

import "sourcegraph.com/sourcegraph/sourcegraph/services/notif"

func init() {
	notif.Disable()
	Init()
}
