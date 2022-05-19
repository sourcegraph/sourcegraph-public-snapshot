package graphql

import gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

type codeIntelIndexerResolver struct {
	name string
	urn  string
}

func (r *codeIntelIndexerResolver) Name() string {
	return r.name
}

func (r *codeIntelIndexerResolver) URL() string {
	if r.urn == "" {
		return ""
	}

	return "https://" + r.urn
}

var (
	lsifNode = codeIntelIndexerResolver{
		name: "lsif-tsc",
		urn:  "github.com/sourcegraph/lsif-node",
	}
	msftNode = codeIntelIndexerResolver{
		name: "msft/lsif-node",
		urn:  "github.com/Microsoft/lsif-node",
	}
	lsifTypescript = codeIntelIndexerResolver{
		name: "lsif-typescript",
		urn:  "github.com/sourcegraph/lsif-typescript",
	}
	lsifJava = codeIntelIndexerResolver{
		name: "scip-java",
		urn:  "github.com/sourcegraph/scip-java",
	}
	msftJava = codeIntelIndexerResolver{
		name: "msft/lsif-java",
		urn:  "github.com/Microsoft/lsif-java",
	}
	lsifGo = codeIntelIndexerResolver{
		name: "lsif-go",
		urn:  "github.com/sourcegraph/lsif-go",
	}
	lsifClang = codeIntelIndexerResolver{
		name: "lsif-clang",
		urn:  "github.com/sourcegraph/lsif-clang",
	}
	lsifCPP = codeIntelIndexerResolver{
		name: "lsif-cpp",
		urn:  "github.com/sourcegraph/lsif-cpp",
	}
	lsifDart = codeIntelIndexerResolver{
		name: "lsif-dart",
		urn:  "github.com/sourcegraph/lsif-dart",
	}
	workivaDart = codeIntelIndexerResolver{
		name: "lsif_indexer",
		urn:  "github.com/Workiva/lsif_indexer",
	}
	hieLSIF = codeIntelIndexerResolver{
		name: "hie-lsif",
		urn:  "github.com/mpickering/hie-lsif",
	}
	lsifJsonnet = codeIntelIndexerResolver{
		name: "lsif-jsonnet",
		urn:  "github.com/sourcegraph/lsif-jsonnet",
	}
	lsifOcaml = codeIntelIndexerResolver{
		name: "lsif-ocaml",
		urn:  "github.com/rvantonder/lsif-ocaml",
	}
	lsifPy = codeIntelIndexerResolver{
		name: "lsif-py",
		urn:  "github.com/sourcegraph/lsif-py",
	}
	rustAnalyzer = codeIntelIndexerResolver{
		name: "rust-analyzer",
		urn:  "github.com/rust-analyzer/rust-analyzer",
	}
	lsifPHP = codeIntelIndexerResolver{
		name: "lsif-php",
		urn:  "github.com/davidrjenni/lsif-php",
	}
	lsifTerraform = codeIntelIndexerResolver{
		name: "lsif-terraform",
		urn:  "github.com/juliosueiras/lsif-terraform",
	}
	lsifDotnet = codeIntelIndexerResolver{
		name: "lsif-dotnet",
		urn:  "github.com/tcz717/LsifDotnet",
	}
)

var allIndexers = []gql.CodeIntelIndexerResolver{
	&lsifNode,
	&msftNode,
	&lsifTypescript,
	&lsifJava,
	&msftJava,
	&lsifGo,
	&lsifClang,
	&lsifCPP,
	&lsifDart,
	&workivaDart,
	&hieLSIF,
	&lsifJsonnet,
	&lsifOcaml,
	&lsifPy,
	&rustAnalyzer,
	&lsifPHP,
	&lsifTerraform,
	&lsifDotnet,
}

// A map of file extension to a list of indexers in order of recommendation
// from most to least.
var languageToIndexer = map[string][]gql.CodeIntelIndexerResolver{
	".go":      {&lsifGo},
	".java":    {&lsifJava, &msftJava},
	".kt":      {&lsifJava},
	".scala":   {&lsifJava},
	".js":      {&lsifTypescript, &lsifNode, &msftNode},
	".jsx":     {&lsifTypescript, &lsifNode, &msftNode},
	".ts":      {&lsifTypescript, &lsifNode, &msftNode},
	".tsx":     {&lsifTypescript, &lsifNode, &msftNode},
	".dart":    {&workivaDart, &lsifDart},
	".c":       {&lsifClang, &lsifCPP},
	".cc":      {&lsifClang, &lsifCPP},
	".cpp":     {&lsifClang, &lsifCPP},
	".cxx":     {&lsifClang, &lsifCPP},
	".h":       {&lsifClang, &lsifCPP},
	".hpp":     {&lsifClang, &lsifCPP},
	".hs":      {&hieLSIF},
	".jsonnet": {&lsifJsonnet},
	".py":      {&lsifPy},
	".ml":      {&lsifOcaml},
	".rs":      {&rustAnalyzer},
	".php":     {&lsifPHP},
	".tf":      {&lsifTerraform},
	".cs":      {&lsifDotnet},
}

var imageToIndexer = map[string]gql.CodeIntelIndexerResolver{
	"sourcegraph/scip-java":       &lsifJava,
	"sourcegraph/lsif-go":         &lsifGo,
	"sourcegraph/lsif-typescript": &lsifTypescript,
	"sourcegraph/lsif-node":       &lsifNode,
	"sourcegraph/lsif-clang":      &lsifClang,
	"davidrjenni/lsif-php":        &lsifPHP,
	"sourcegraph/lsif-rust":       &rustAnalyzer,
}
