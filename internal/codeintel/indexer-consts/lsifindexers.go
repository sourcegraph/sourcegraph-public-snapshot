package indexerconsts

import gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

type codeIntelIndexerResolver struct {
	name, urn string
}

func (r *codeIntelIndexerResolver) Name() string {
	return r.name
}

func (r *codeIntelIndexerResolver) URL() string {
	return "https://" + r.urn
}

var (
	LSIFNode = codeIntelIndexerResolver{
		name: "lsif-tsc",
		urn:  "github.com/sourcegraph/lsif-node",
	}
	MSFTNode = codeIntelIndexerResolver{
		name: "msft/lsif-node",
		urn:  "github.com/Microsoft/lsif-node",
	}
	LSIFTypescript = codeIntelIndexerResolver{
		name: "lsif-typescript",
		urn:  "github.com/sourcegraph/lsif-typescript",
	}
	LSIFJava = codeIntelIndexerResolver{
		name: "lsif-java",
		urn:  "github.com/sourcegraph/lsif-java",
	}
	MSFTJava = codeIntelIndexerResolver{
		name: "msft/lsif-java",
		urn:  "github.com/Microsoft/lsif-java",
	}
	LSIFGo = codeIntelIndexerResolver{
		name: "lsif-go",
		urn:  "github.com/sourcegraph/lsif-go",
	}
	LSIFClang = codeIntelIndexerResolver{
		name: "lsif-clang",
		urn:  "github.com/sourcegraph/lsif-clang",
	}
	LSIFCpp = codeIntelIndexerResolver{
		name: "lsif-cpp",
		urn:  "github.com/sourcegraph/lsif-cpp",
	}
	LSIFDart = codeIntelIndexerResolver{
		name: "lsif-dart",
		urn:  "github.com/sourcegraph/lsif-dart",
	}
	WorkivaDart = codeIntelIndexerResolver{
		name: "lsif_indexer",
		urn:  "github.com/Workiva/lsif_indexer",
	}
	HieLSIF = codeIntelIndexerResolver{
		name: "hie-lsif",
		urn:  "github.com/mpickering/hie-lsif",
	}
	LSIFJsonnet = codeIntelIndexerResolver{
		name: "lsif-jsonnet",
		urn:  "github.com/sourcegraph/lsif-jsonnet",
	}
	LSIFOcaml = codeIntelIndexerResolver{
		name: "lsif-ocaml",
		urn:  "github.com/rvantonder/lsif-ocaml",
	}
	LSIFPy = codeIntelIndexerResolver{
		name: "lsif-py",
		urn:  "github.com/sourcegraph/lsif-py",
	}
	RustAnalyzer = codeIntelIndexerResolver{
		name: "rust-analyzer",
		urn:  "github.com/rust-analyzer/rust-analyzer",
	}
	LSIFPhp = codeIntelIndexerResolver{
		name: "lsif-php",
		urn:  "github.com/davidrjenni/lsif-php",
	}
	LSIFTerraform = codeIntelIndexerResolver{
		name: "lsif-terraform",
		urn:  "github.com/juliosueiras/lsif-terraform",
	}
	LSIFDotnet = codeIntelIndexerResolver{
		name: "lsif-dotnet",
		urn:  "github.com/tcz717/LsifDotnet",
	}
)

// A map of file extension to a list of indexers in order of recommendation
// from most to least.
var LanguageToIndexer = map[string][]gql.CodeIntelIndexerResolver{
	".go":      {&LSIFGo},
	".java":    {&LSIFJava, &MSFTJava},
	".kt":      {&LSIFJava},
	".scala":   {&LSIFJava},
	".js":      {&LSIFTypescript, &LSIFNode, &MSFTNode},
	".jsx":     {&LSIFTypescript, &LSIFNode, &MSFTNode},
	".ts":      {&LSIFTypescript, &LSIFNode, &MSFTNode},
	".tsx":     {&LSIFTypescript, &LSIFNode, &MSFTNode},
	".dart":    {&WorkivaDart, &LSIFDart},
	".c":       {&LSIFClang, &LSIFCpp},
	".cc":      {&LSIFClang, &LSIFCpp},
	".cpp":     {&LSIFClang, &LSIFCpp},
	".cxx":     {&LSIFClang, &LSIFCpp},
	".h":       {&LSIFClang, &LSIFCpp},
	".hs":      {&HieLSIF},
	".jsonnet": {&LSIFJsonnet},
	".py":      {&LSIFPy},
	".ml":      {&LSIFOcaml},
	".rs":      {&RustAnalyzer},
	".php":     {&LSIFPhp},
	".tf":      {&LSIFTerraform},
	".cs":      {&LSIFDotnet},
}
