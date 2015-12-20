package app

import (
	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
)

func showSearchForm(tmpl *tmpl.Common) bool {
	if appconf.Flags.DisableSearch {
		return false
	}

	if appconf.Flags.CustomNavLayout != "" {
		return false
	}

	return tmpl.TemplateName != "error/error.html" && tmpl.CurrentRouteVars["Repo"] != ""
}
