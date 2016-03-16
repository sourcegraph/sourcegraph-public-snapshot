package app

import (
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
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
