package app

import "src.sourcegraph.com/sourcegraph/notif"

func init() {
	notif.MustBeDisabled()
	Init()
}
