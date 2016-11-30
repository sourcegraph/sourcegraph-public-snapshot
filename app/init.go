package app

import (
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
)

func Init() {
	tmpl.LoadOnce()
}
