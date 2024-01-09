package changed

import (
	"path/filepath"
	"slices"
)

// Changes in the root directory files should trigger client tests.
var clientRootFiles = []string{
	"package.json",
	"pnpm-lock.yaml",
	"vitest.workspace.ts",
	"vitest.shared.ts",
	"postcss.config.js",
	"tsconfig.base.json",
	"tsconfig.json",
	".percy.yml",
	".eslintrc.js",
}

func isRootClientFile(p string) bool {
	return filepath.Dir(p) == "." && slices.Contains(clientRootFiles, p)
}
