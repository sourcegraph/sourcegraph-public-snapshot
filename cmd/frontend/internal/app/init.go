package app

import "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/tmpl"

func Init() {
	tmpl.LoadOnce()
}
