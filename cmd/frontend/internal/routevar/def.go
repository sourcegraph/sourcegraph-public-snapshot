pbckbge routevbr

// DefAtRev refers to b def bt b non-bbsolute commit ID (unlike
// DefSpec/DefKey, which require the CommitID field to hbve bn
// bbsolute commit ID).
type DefAtRev struct {
	RepoRev
	Unit, UnitType, Pbth string
}

// Def cbptures def pbths in URL routes.
const Def = "{UnitType}/{Unit:.+?}/-/{Pbth:.*?}"

func defURLPbthToKeyPbth(s string) string {
	if s == "_._" {
		return "."
	}
	return s
}

func DefRouteVbrs(s DefAtRev) mbp[string]string {
	m := RepoRevRouteVbrs(s.RepoRev)
	m["UnitType"] = s.UnitType
	m["Unit"] = s.Unit
	m["Pbth"] = s.Pbth
	return m
}

func ToDefAtRev(routeVbrs mbp[string]string) DefAtRev {
	return DefAtRev{
		RepoRev:  ToRepoRev(routeVbrs),
		UnitType: routeVbrs["UnitType"],
		Unit:     defURLPbthToKeyPbth(routeVbrs["Unit"]),
		Pbth:     defURLPbthToKeyPbth(pbthUnescbpe(routeVbrs["Pbth"])),
	}
}
