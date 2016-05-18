package python

import (
	"log"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

const (
	DistPackageSourceUnitType = "PipPackage"
)

// Format outputted by scanner
type pkgInfo struct {
	RootDir     string   `json:"rootdir,omitempty"`
	ProjectName string   `json:"project_name,omitempty"`
	Version     string   `json:"version,omitempty"`
	RepoURL     string   `json:"repo_url,omitempty"`
	Packages    []string `json:"packages,omitempty"`
	Modules     []string `json:"modules,omitempty"`
	Scripts     []string `json:"scripts,omitempty"`
	Author      string   `json:"author,omitempty"`
	Description string   `json:"description,omitempty"`
}

func (p *pkgInfo) SourceUnit() *unit.SourceUnit {
	repoURI, err := graph.TryMakeURI(p.RepoURL)
	if err != nil {
		log.Printf("Could not make repo URI from %s: %s", p.RepoURL, err)
		repoURI = ""
	}
	return &unit.SourceUnit{
		Key: unit.Key{
			Name: p.ProjectName,
			Type: DistPackageSourceUnitType,
			Repo: repoURI,
		},
		Info: unit.Info{
			Dir:          p.RootDir,
			Dependencies: nil, // nil, because scanner does not resolve dependencies
			Ops:          map[string][]byte{"depresolve": nil, "graph": nil},
		},
	}
}

type requirement struct {
	ProjectName string      `json:"project_name"`
	UnsafeName  string      `json:"unsafe_name"`
	Key         string      `json:"key"`
	Specs       [][2]string `json:"specs"`
	Extras      []string    `json:"extras"`
	RepoURL     string      `json:"repo_url"`
	Packages    []string    `json:"packages"`
	Modules     []string    `json:"modules"`
	Resolved    bool        `json:"resolved"`
	Type        string      `json:"type"`
}

func (r *requirement) SourceUnit() *unit.SourceUnit {
	return &unit.SourceUnit{
		Key: unit.Key{
			Name: r.ProjectName,
			Type: DistPackageSourceUnitType,
			Repo: graph.MakeURI(r.RepoURL),
		},
	}
}
