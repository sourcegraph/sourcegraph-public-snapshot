package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	goEnterpriseImport = lint.RunScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh")
	tsEnterpriseImport = lint.RunScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh")
)
