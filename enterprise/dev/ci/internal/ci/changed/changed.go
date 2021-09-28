package changed

import (
	"strings"
)

type ChangedFiles []string

// affectsDocs returns whether the ChangedFiles affects documentation.
func (c ChangedFiles) AffectsDocs() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "doc/") && p != "CHANGELOG.md" {
			return true
		}
	}
	return false
}

// affectsSg returns whether the ChangedFiles affects the ./dev/sg folder.
func (c ChangedFiles) AffectsSg() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "dev/sg/") {
			return true
		}
	}
	return false
}

// affectsGo returns whether the ChangedFiles affects go files.
func (c ChangedFiles) AffectsGo() bool {
	for _, p := range c {
		if strings.HasSuffix(p, ".go") || p == "go.sum" || p == "go.mod" {
			return true
		}
	}
	return false
}

// affectsDockerfiles returns whether the ChangedFiles affects Dockerfiles.
func (c ChangedFiles) AffectsDockerfiles() bool {
	for _, p := range c {
		if strings.HasPrefix(p, "Dockerfile") || strings.HasSuffix(p, "Dockerfile") {
			return true
		}
	}
	return false
}

// affectsGraphQL returns whether the ChangedFiles affects GraphQL files
func (c ChangedFiles) AffectsGraphQL() bool {
	for _, p := range c {
		if strings.HasSuffix(p, ".graphql") {
			return true
		}
	}
	return false
}

// Check if files that affect client code were changed. Used to detect if we need to run Puppeteer or Chromatic tests.
func (c ChangedFiles) AffectsClient() bool {
	for _, p := range c {
		if !strings.HasSuffix(p, ".md") && (strings.HasPrefix(p, "client/") || isAllowedRootFile(p)) {
			return true
		}
	}
	return false
}
