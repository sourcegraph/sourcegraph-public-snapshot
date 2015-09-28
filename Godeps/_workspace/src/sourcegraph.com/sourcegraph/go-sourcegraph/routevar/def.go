package routevar

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"
)

// Def captures def paths in URL routes.
//
// We want the def routes to match the 2 following forms:
//
//   1. /.MyUnitType/.def/MyDef (i.e., Unit == ".")
//   2. /.MyUnitType/MyUnitPath1/.def/MyDef (i.e., Unit == "MyUnitPath1")
//
// To achieve this, we use a non-picky regexp for rawUnit and then sort it
// out in the FixDefUnitVars PostMatchFunc.
var Def = `.{UnitType}/{rawUnit:.*}.def{Path:(?:(?:/(?:[^/.][^/]*/)*(?:[^/.][^/]*))|)}`

// FixDefUnitVars is a mux.PostMatchFunc that cleans up the dummy
// rawUnit route variable matched by Def. See the docs for Def for
// more information.
func FixDefUnitVars(req *http.Request, match *mux.RouteMatch, r *mux.Route) {
	match.Vars["Path"] = strings.TrimPrefix(match.Vars["Path"], "/")
	if path := match.Vars["Path"]; path == "" {
		match.Vars["Path"] = "."
	}
	match.Vars["Path"] = pathUnescape(match.Vars["Path"])

	rawUnit := match.Vars["rawUnit"]
	if rawUnit == "" {
		match.Vars["Unit"] = "."
	} else {
		match.Vars["Unit"] = strings.TrimSuffix(rawUnit, "/")
	}
	delete(match.Vars, "rawUnit")
}

// PrepareDefRouteVars is a mux.BuildVarsFunc that converts from a "Unit"
// route variable to the dummy "rawUnit" route variable that actually appears in
// the route regexp pattern.
func PrepareDefRouteVars(vars map[string]string) map[string]string {
	if path := vars["Path"]; path == "." {
		vars["Path"] = ""
	} else if path != "" {
		vars["Path"] = "/" + path
	}

	vars["Path"] = pathEscape(vars["Path"])

	if unit := vars["Unit"]; unit == "." {
		vars["rawUnit"] = ""
	} else {
		vars["rawUnit"] = unit + "/"
	}
	delete(vars, "Unit")

	return vars
}

// pathEscape is a limited version of url.QueryEscape that only escapes '?'.
func pathEscape(p string) string {
	return strings.Replace(p, "?", "%3F", -1)
}

// pathUnescape is a limited version of url.QueryEscape that only unescapes '?'.
func pathUnescape(p string) string {
	return strings.Replace(p, "%3F", "?", -1)
}
