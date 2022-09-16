package changed

import (
	"bytes"
	"os"
	"strings"
)

type Diff uint32

const (
	// None indicates no diff. Use sparingly.
	None Diff = 0

	Go Diff = 1 << iota
	Client
	GraphQL
	DatabaseSchema
	Docs
	Dockerfiles
	ExecutorDockerRegistryMirror
	CIScripts
	Terraform
	SVG
	Shell
	DockerImages

	// All indicates all changes should be considered included in this diff, except None.
	All
)

// ForEachDiffType iterates all Diff types except None and All and calls the callback on
// each.
func ForEachDiffType(callback func(d Diff)) {
	const firstDiffType = Diff(1 << 1)
	for d := firstDiffType; d < All; d <<= 1 {
		callback(d)
	}
}

// topLevelGoDirs is a slice of directories which contain most of our go code.
// A PR could just mutate test data or embedded files, so we treat any change
// in these directories as a go change.
var topLevelGoDirs = []string{
	"cmd",
	"enterprise/cmd",
	"enterprise/internal",
	"internal",
	"lib",
	"migrations",
	"monitoring",
	"schema",
}

// ParseDiff identifies what has changed in files by generating a Diff that can be used
// to check for specific changes, e.g.
//
//	if diff.Has(changed.Client | changed.GraphQL) { ... }
//
// To introduce a new type of Diff, add it a new Diff constant above and add a check in
// this function to identify the Diff.
func ParseDiff(files []string) (diff Diff) {
	for _, p := range files {
		// Affects Go
		if strings.HasSuffix(p, ".go") || p == "go.sum" || p == "go.mod" {
			diff |= Go
		}
		if strings.HasSuffix(p, "dev/ci/go-test.sh") {
			diff |= Go
		}
		for _, dir := range topLevelGoDirs {
			if strings.HasPrefix(p, dir+"/") {
				diff |= Go
			}
		}
		if p == "sg.config.yaml" {
			// sg config affects generated output and potentially tests and checks that we
			// run in the future, so we consider this to have affected Go.
			diff |= Go
		}

		// Client
		if !strings.HasSuffix(p, ".md") && (isRootClientFile(p) || strings.HasPrefix(p, "client/")) {
			diff |= Client
		}
		if strings.HasSuffix(p, "dev/ci/yarn-test.sh") {
			diff |= Client
		}
		// dev/release contains a nodejs script that doesn't have tests but needs to be
		// linted with Client linters
		if strings.HasPrefix(p, "dev/release/") {
			diff |= Client
		}

		// Affects GraphQL
		if strings.HasSuffix(p, ".graphql") {
			diff |= GraphQL
		}

		// Affects DB schema
		if strings.HasPrefix(p, "migrations/") {
			diff |= (DatabaseSchema | Go)
		}
		if strings.HasPrefix(p, "dev/ci/go-backcompat") {
			diff |= DatabaseSchema
		}

		// Affects docs
		if strings.HasPrefix(p, "doc/") || strings.HasSuffix(p, ".md") {
			diff |= Docs
		}
		if strings.HasSuffix(p, ".yaml") || strings.HasSuffix(p, ".yml") {
			diff |= Docs
		}
		if strings.HasSuffix(p, ".json") || strings.HasSuffix(p, ".jsonc") || strings.HasSuffix(p, ".json5") {
			diff |= Docs
		}

		// Affects Dockerfiles (which assumes images are being changed as well)
		if strings.HasPrefix(p, "Dockerfile") || strings.HasSuffix(p, "Dockerfile") {
			diff |= (Dockerfiles | DockerImages)
		}
		// Affects anything in docker-images directories (which implies image build
		// scripts and/or resources are affected)
		if strings.HasPrefix(p, "docker-images/") {
			diff |= DockerImages
		}

		// Affects executor docker registry mirror
		if strings.HasPrefix(p, "enterprise/cmd/executor/docker-mirror/") {
			diff |= ExecutorDockerRegistryMirror
		}

		// Affects CI scripts
		if strings.HasPrefix(p, "enterprise/dev/ci/scripts") {
			diff |= CIScripts
		}

		// Affects Terraform
		if strings.HasSuffix(p, ".tf") {
			diff |= Terraform
		}

		// Affects SVG files
		if strings.HasSuffix(p, ".svg") {
			diff |= SVG
		}

		// Affects scripts
		if strings.HasSuffix(p, ".sh") {
			diff |= Shell
		}
		// Read the file to check if it is secretly a shell script
		f, err := os.Open(p)
		if err == nil {
			b := make([]byte, 19) // "#!/usr/bin/env bash" = 19 chars
			_, _ = f.Read(b)
			if bytes.Compare(b[0:2], []byte("#!")) == 0 && bytes.Contains(b, []byte("bash")) {
				// If the file starts with a shebang and has "bash" somewhere after, it's most probably
				// some shell script.
				diff |= Shell
			}
			// Close the file immediately - we don't want to defer, this loop can go for
			// quite a while.
			f.Close()
		}
	}
	return
}

func (d Diff) String() string {
	switch d {
	case None:
		return "None"

	case Go:
		return "Go"
	case Client:
		return "Client"
	case GraphQL:
		return "GraphQL"
	case DatabaseSchema:
		return "DatabaseSchema"
	case Docs:
		return "Docs"
	case Dockerfiles:
		return "Dockerfiles"
	case ExecutorDockerRegistryMirror:
		return "ExecutorDockerRegistryMirror"
	case CIScripts:
		return "CIScripts"
	case Terraform:
		return "Terraform"
	case SVG:
		return "SVG"
	case Shell:
		return "Shell"
	case DockerImages:
		return "DockerImages"

	case All:
		return "All"
	}

	var allDiffs []string
	ForEachDiffType(func(checkDiff Diff) {
		diffName := checkDiff.String()
		if diffName != "" && d.Has(checkDiff) {
			allDiffs = append(allDiffs, diffName)
		}
	})
	return strings.Join(allDiffs, ", ")
}

// Has returns true if d has the target diff.
func (d Diff) Has(target Diff) bool {
	switch d {
	case None:
		// If None, the only other Diff type that matches this is another None.
		return target == None

	case All:
		// If All, this change includes all other Diff types except None.
		return target != None

	default:
		return d&target != 0
	}
}
