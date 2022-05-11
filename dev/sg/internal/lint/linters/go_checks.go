package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	goFmt          = lint.RunScript("Go format", "dev/check/gofmt.sh")
	goLint         = lint.RunScript("Go lint", "dev/check/go-lint.sh")
	goDBConnImport = lint.RunScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh")
)
