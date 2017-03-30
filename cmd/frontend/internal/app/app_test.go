package app

import "sourcegraph.com/sourcegraph/sourcegraph/pkg/notif"

func init() {
	notif.Disable()
	Init()
}
