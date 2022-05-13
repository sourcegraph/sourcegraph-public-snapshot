package linters

import "github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"

var (
	inlineTemplates = lint.RunScript("Inline templates", "dev/check/template-inlines.sh")
)
