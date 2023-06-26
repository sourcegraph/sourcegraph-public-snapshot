//go:build !windows
// +build !windows

package singleprogram

func startEmbeddedPostgreSQL(logger log.Logger, pgRootDir string) (StopPostgresFunc, *postgresqlEnvVars, error) {
	// Note: some linux distributions (eg NixOS) do not ship with the dynamic
	// linker at the "standard" location which the embedded postgres
	// executables rely on. Give a nice error instead of the confusing "file
	// not found" error.
	//
	// We could consider extending embedded-postgres to use something like
	// patchelf, but this is non-trivial.
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		ldso := "/lib64/ld-linux-x86-64.so.2"
		if _, err := os.Stat(ldso); err != nil {
			return noopStop, nil, errors.Errorf("could not use embedded-postgres since %q is missing - see https://github.com/sourcegraph/sourcegraph/issues/52360 for more details", ldso)
		}
	}

	// Note: on macOS unix socket paths must be <103 bytes, so we place them in the home directory.
	current, err := user.Current()
	if err != nil {
		return noopStop, nil, errors.Wrap(err, "user.Current")
	}
	unixSocketDir := filepath.Join(current.HomeDir, ".sourcegraph-psql")
	err = os.RemoveAll(unixSocketDir)
	if err != nil {
		logger.Warn("unable to remove previous connection", log.Error(err))
	}
	err = os.MkdirAll(unixSocketDir, os.ModePerm)
	if err != nil {
		return noopStop, nil, errors.Wrap(err, "creating unix socket dir")
	}

	vars := &postgresqlEnvVars{
		PGPORT:       "",
		PGHOST:       unixSocketDir,
		PGUSER:       "sourcegraph",
		PGPASSWORD:   "",
		PGDATABASE:   "sourcegraph",
		PGSSLMODE:    "disable",
		PGDATASOURCE: "postgresql:///sourcegraph?host=" + unixSocketDir,
	}

	db := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Version(embeddedpostgres.V14).
			BinariesPath(filepath.Join(pgRootDir, "bin")).
			DataPath(filepath.Join(pgRootDir, "data")).
			RuntimePath(filepath.Join(pgRootDir, "runtime")).
			Username(vars.PGUSER).
			Database(vars.PGDATABASE).
			UseUnixSocket(unixSocketDir).
			StartTimeout(120 * time.Second).
			Logger(debugLogLinesWriter(logger, "postgres output line")),
	)
	if err := db.Start(); err != nil {
		return noopStop, nil, err
	}

	return db.Stop, vars, nil
}