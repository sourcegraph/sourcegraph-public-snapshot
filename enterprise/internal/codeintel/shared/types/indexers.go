package types

type CodeIntelIndexer struct {
	Name string
	URN  string
}

var (
	lsifNode = CodeIntelIndexer{
		Name: "lsif-tsc",
		URN:  "github.com/sourcegraph/lsif-node",
	}
	msftNode = CodeIntelIndexer{
		Name: "msft/lsif-node",
		URN:  "github.com/Microsoft/lsif-node",
	}
	scipTypescript = CodeIntelIndexer{
		Name: "scip-typescript",
		URN:  "github.com/sourcegraph/scip-typescript",
	}
	scipJava = CodeIntelIndexer{
		Name: "scip-java",
		URN:  "github.com/sourcegraph/scip-java",
	}
	msftJava = CodeIntelIndexer{
		Name: "msft/lsif-java",
		URN:  "github.com/Microsoft/lsif-java",
	}
	lsifGo = CodeIntelIndexer{
		Name: "lsif-go",
		URN:  "github.com/sourcegraph/lsif-go",
	}
	lsifClang = CodeIntelIndexer{
		Name: "lsif-clang",
		URN:  "github.com/sourcegraph/lsif-clang",
	}
	lsifCPP = CodeIntelIndexer{
		Name: "lsif-cpp",
		URN:  "github.com/sourcegraph/lsif-cpp",
	}
	lsifDart = CodeIntelIndexer{
		Name: "lsif-dart",
		URN:  "github.com/sourcegraph/lsif-dart",
	}
	workivaDart = CodeIntelIndexer{
		Name: "lsif_indexer",
		URN:  "github.com/Workiva/lsif_indexer",
	}
	hieLSIF = CodeIntelIndexer{
		Name: "hie-lsif",
		URN:  "github.com/mpickering/hie-lsif",
	}
	lsifJsonnet = CodeIntelIndexer{
		Name: "lsif-jsonnet",
		URN:  "github.com/sourcegraph/lsif-jsonnet",
	}
	lsifOcaml = CodeIntelIndexer{
		Name: "lsif-ocaml",
		URN:  "github.com/rvantonder/lsif-ocaml",
	}
	scipPython = CodeIntelIndexer{
		Name: "scip-python",
		URN:  "github.com/sourcegraph/scip-python",
	}
	rustAnalyzer = CodeIntelIndexer{
		Name: "rust-analyzer",
		URN:  "github.com/rust-analyzer/rust-analyzer",
	}
	lsifPHP = CodeIntelIndexer{
		Name: "lsif-php",
		URN:  "github.com/davidrjenni/lsif-php",
	}
	lsifTerraform = CodeIntelIndexer{
		Name: "lsif-terraform",
		URN:  "github.com/juliosueiras/lsif-terraform",
	}
	lsifDotnet = CodeIntelIndexer{
		Name: "lsif-dotnet",
		URN:  "github.com/tcz717/LsifDotnet",
	}
)

var AllIndexers = []CodeIntelIndexer{
	lsifNode,
	msftNode,
	scipTypescript,
	scipJava,
	msftJava,
	lsifGo,
	lsifClang,
	lsifCPP,
	lsifDart,
	workivaDart,
	hieLSIF,
	lsifJsonnet,
	lsifOcaml,
	scipPython,
	rustAnalyzer,
	lsifPHP,
	lsifTerraform,
	lsifDotnet,
}

// A map of file extension to a list of indexers in order of recommendation
// from most to least.
var LanguageToIndexer = map[string][]CodeIntelIndexer{
	".go":      {lsifGo},
	".java":    {scipJava, msftJava},
	".kt":      {scipJava},
	".scala":   {scipJava},
	".js":      {scipTypescript, lsifNode, msftNode},
	".jsx":     {scipTypescript, lsifNode, msftNode},
	".ts":      {scipTypescript, lsifNode, msftNode},
	".tsx":     {scipTypescript, lsifNode, msftNode},
	".dart":    {workivaDart, lsifDart},
	".c":       {lsifClang, lsifCPP},
	".cc":      {lsifClang, lsifCPP},
	".cpp":     {lsifClang, lsifCPP},
	".cxx":     {lsifClang, lsifCPP},
	".h":       {lsifClang, lsifCPP},
	".hpp":     {lsifClang, lsifCPP},
	".hs":      {hieLSIF},
	".jsonnet": {lsifJsonnet},
	".py":      {scipPython},
	".ml":      {lsifOcaml},
	".rs":      {rustAnalyzer},
	".php":     {lsifPHP},
	".tf":      {lsifTerraform},
	".cs":      {lsifDotnet},
}

var ImageToIndexer = map[string]CodeIntelIndexer{
	"sourcegraph/scip-java":       scipJava,
	"sourcegraph/lsif-go":         lsifGo,
	"sourcegraph/scip-typescript": scipTypescript,
	"sourcegraph/lsif-node":       lsifNode,
	"sourcegraph/lsif-clang":      lsifClang,
	"davidrjenni/lsif-php":        lsifPHP,
	"sourcegraph/lsif-rust":       rustAnalyzer,
	"sourcegraph/scip-python":     scipPython,
}

var PreferredIndexers = map[string]CodeIntelIndexer{
	"lsif-node":       scipTypescript,
	"lsif-tsc":        scipTypescript,
	"scip-typescript": scipTypescript,
	"scip-java":       scipJava,
	"lsif-java":       scipJava,
	"lsif-go":         lsifGo,
	"lsif-clang":      lsifClang,
	"lsif-cpp":        lsifCPP,
	"lsif-dart":       lsifDart,
	"hie-lsif":        hieLSIF,
	"lsif-jsonnet":    lsifJsonnet,
	"lsif-ocaml":      lsifOcaml,
	"scip-python":     scipPython,
	"lsif-rust":       rustAnalyzer,
	"rust-analyzer":   rustAnalyzer,
	"lsif-php":        lsifPHP,
	"lsif-terraform":  lsifTerraform,
	"lsif-dotnet":     lsifDotnet,
}
