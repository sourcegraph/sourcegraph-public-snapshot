package changed

import (
	"path/filepath"
	"slices"
)

var pnpmFiles = []string{
	"package.json",
	"pnpm-lock.yaml",
	"pnpm-workspace.yaml",
}

// Changes in the root directory files should trigger client tests.
var clientRootFiles = append(pnpmFiles, []string{
	"vitest.workspace.ts",
	"vitest.shared.ts",
	"postcss.config.js",
	"tsconfig.base.json",
	"tsconfig.json",
	".eslintrc.js",
}...)

func isRootClientFile(p string) bool {
	return filepath.Dir(p) == "." && slices.Contains(clientRootFiles, p)
}
