package changed

import (
	"path/filepath"
	"strings"
)

// Files is the list of changed files to operate over in a pipeline.
//
// Helper functions on Files should all be in the format `AffectsXYZ`.
type Files []string

// AffectsPathsPrefixedBy returns whether the chanegs affects CI scripts.
func (f Files) AffectsCIScripts() bool {
	for _, p := range f {
		if strings.HasPrefix(p, "enterprise/dev/ci/scripts") {
			return true
		}
	}
	return false
}

// AffectsDocs returns whether the changes affects documentation.
func (f Files) AffectsDocs() bool {
	for _, p := range f {
		if strings.HasPrefix(p, "doc/") && p != "CHANGELOG.md" {
			return true
		}
	}
	return false
}

// affectsSg returns whether the changes affects the ./dev/sg folder.
func (c Files) AffectsSg() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "dev/sg/") {
			return true
		}
	}
	return false
}

// AffectsTerraformFiles returns whether the changes affects terraform files.
// This will be based on changes to TF files for now.
func (c Files) AffectsTerraformFiles() bool {
	for _, p := range c {
		if strings.HasSuffix(p, ".tf") {
			return true
		}
	}
	return false
}

// AffectsFilesWithExt returns whether the changes affects files with the given extension.
// Extension must be passed with the dot, eg: ".svg"
func (c Files) AffectsFilesWithExt(ext string) bool {
	for _, p := range c {
		if strings.HasSuffix(p, ext) {
			return true
		}
	}
	return false
}

// AffectsGo returns whether the changes affects go files.
func (c Files) AffectsGo() bool {
	for _, p := range c {
		if strings.HasSuffix(p, ".go") || p == "go.sum" || p == "go.mod" {
			return true
		}
	}
	return c.AffectsDatabaseSchema()
}

// AffectsDockerfiles returns whether the changes affects Dockerfiles.
func (f Files) AffectsDockerfiles() bool {
	for _, p := range f {
		if strings.HasPrefix(p, "Dockerfile") || strings.HasSuffix(p, "Dockerfile") {
			return true
		}
	}
	return false
}

// AffectsDatabaseSchema returns whether the changes affect the database schema definition.
func (c Files) AffectsDatabaseSchema() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "migrations/") {
			return true
		}
	}
	return false
}

// AffectsGraphQL returns whether the changes affects GraphQL files
func (f Files) AffectsGraphQL() bool {
	for _, p := range f {
		if strings.HasSuffix(p, ".graphql") {
			return true
		}
	}
	return false
}

// AffectsClient returns whether files that affect client code were changed.
// Used to detect if we need to run Puppeteer or Chromatic tests.
func (f Files) AffectsClient() bool {
	isRootClientFile := func(p string) bool {
		return filepath.Dir(p) == "." && contains(clientRootFiles, p)
	}
	for _, p := range f {
		if !strings.HasSuffix(p, ".md") && (isRootClientFile(p) || strings.HasPrefix(p, "client/")) {
			return true
		}
	}
	return false
}

// AffectsExecutorDockerRegistryMirror returns whether files that affect the executor
// docker registry mirror were changed.
func (f Files) AffectsExecutorDockerRegistryMirror() bool {
	for _, p := range f {
		if strings.HasPrefix(p, "enterprise/cmd/executor/docker-mirror/") {
			return true
		}
	}
	return false
}
