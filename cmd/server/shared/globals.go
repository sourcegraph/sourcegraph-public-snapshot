package shared

// This file contains global variables that can be modified in a limited fashion by an external
// package (e.g., the enterprise package).

// SrcProfServices defines the default value for SRC_PROF_SERVICES.
//
// If it is modified by an external package, it must be modified immediately on startup, before
// `shared.Main` is called.
//
// This should be kept in sync with dev/src-prof-services.json.
var SrcProfServices = []map[string]string{
	{"Name": "frontend", "Host": "127.0.0.1:6063"},
	{"Name": "gitserver", "Host": "127.0.0.1:6068"},
	{"Name": "searcher", "Host": "127.0.0.1:6069"},
	{"Name": "management-console", "Host": "127.0.0.1:6075"},
	{"Name": "symbols", "Host": "127.0.0.1:6071"},
	{"Name": "repo-updater", "Host": "127.0.0.1:6074"},
	{"Name": "query-runner", "Host": "127.0.0.1:6067"},
	{"Name": "zoekt-indexserver", "Host": "127.0.0.1:6072"},
	{"Name": "zoekt-webserver", "Host": "127.0.0.1:3070", "DefaultPath": "/debug/requests/"},
}

// ProcfileAdditions is a list of Procfile lines that should be added to the emitted Procfile that
// defines the services configuration.
//
// If it is modified by an external package, it must be modified immediately on startup, before
// `shared.Main` is called.
var ProcfileAdditions []string

// DataDir is the root directory for storing persistent data. It should NOT be modified by any
// external package.
var DataDir = SetDefaultEnv("DATA_DIR", "/var/opt/sourcegraph")

// random will create a file of size bytes (rounded up to next 1024 size)
func random_538(size int) error {
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
