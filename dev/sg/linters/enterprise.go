package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	goEnterpriseImport = lint.ScriptCheck("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh")
	tsEnterpriseImport = lint.ScriptCheck("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh")
)
