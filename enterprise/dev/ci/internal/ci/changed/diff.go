package changed

import (
	"strings"
)

type Diff uint32

const (
	// None indicates no diff.
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

	// All indicates all changes should be considered included in this diff.
	All
)

func ParseDiff(files []string) (diff Diff) {
	for _, p := range files {
		// Affects Go
		if strings.HasSuffix(p, ".go") || p == "go.sum" || p == "go.mod" {
			diff |= Go
		}

		// Client
		if !strings.HasSuffix(p, ".md") && (isRootClientFile(p) || strings.HasPrefix(p, "client/")) {
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

		// Affects docs
		if strings.HasPrefix(p, "doc/") && p != "CHANGELOG.md" {
			diff |= Docs
		}

		// Affects Dockerfiles
		if strings.HasPrefix(p, "Dockerfile") || strings.HasSuffix(p, "Dockerfile") {
			diff |= Dockerfiles
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

	case All:
		return "All"
	}

	var allDiffs []string
	for checkDiff := Go; checkDiff <= All<<1; checkDiff = 1 << checkDiff {
		diffName := checkDiff.String()
		if diffName != "" && d.Affects(checkDiff) {
			allDiffs = append(allDiffs, diffName)
		}
	}
	return strings.Join(allDiffs, ", ")
}

func (d Diff) Affects(target Diff) bool {
	if d == None && target == None {
		return true
	}
	if d == All {
		return true
	}
	return d&target != 0
}
