package routevar

// DefAtRev refers to a def at a non-absolute commit ID (unlike
// DefSpec/DefKey, which require the CommitID field to have an
// absolute commit ID).
type DefAtRev struct {
	RepoRev
	Unit, UnitType, Path string
}

// Def captures def paths in URL routes.
const Def = "{UnitType}/{Unit:.+?}/-/{Path:.*?}"

func defURLPathToKeyPath(s string) string {
	if s == "_._" {
		return "."
	}
	return s
}

func DefRouteVars(s DefAtRev) map[string]string {
	m := RepoRevRouteVars(s.RepoRev)
	m["UnitType"] = s.UnitType
	m["Unit"] = s.Unit
	m["Path"] = s.Path
	return m
}

func ToDefAtRev(routeVars map[string]string) DefAtRev {
	return DefAtRev{
		RepoRev:  ToRepoRev(routeVars),
		UnitType: routeVars["UnitType"],
		Unit:     defURLPathToKeyPath(routeVars["Unit"]),
		Path:     defURLPathToKeyPath(pathUnescape(routeVars["Path"])),
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_881(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
