package types

import "fmt"

type CodeIntelIndexer struct {
	LanguageKey  string
	Name         string
	URN          string
	DockerImages []string
}

// AllIndexers is a list of all detectable/suggested indexers known to Sourcegraph.
// Two indexers with the same language key will be preferred according to the given order.
var AllIndexers = []CodeIntelIndexer{
	// C/C++
	makeInternalIndexer("C++", "scip-clang"),
	makeInternalIndexer("C++", "lsif-clang"),
	makeInternalIndexer("C++", "lsif-cpp"),

	// Dart
	makeInternalIndexer("Dart", "lsif-dart"),
	makeIndexer("Dart", "lsif_indexer", "github.com/Workiva/lsif_indexer"),

	// C#, F# (.xls wtf :screamcat:)
	makeInternalIndexer("DotNet", "scip-dotnet"),
	makeIndexer("DotNet", "lsif-dotnet", "github.com/tcz717/LsifDotnet"),

	// Go
	makeInternalIndexer("Go", "scip-go"),
	makeInternalIndexer("Go", "lsif-go"),

	// HIE
	makeIndexer("HIE", "hie-lsif", " github.com/mpickering/hie-lsif"),

	// Jsonnet
	makeInternalIndexer("Jsonnet", "lsif-jsonnet"),

	// Java/Kotlin/Scala
	makeInternalIndexer("JVM", "scip-java"),
	makeInternalIndexer("JVM", "lsif-java"),

	// OCaml
	makeIndexer("OCaml", "lsif-ocaml", "github.com/rvantonder/lsif-ocaml"),

	// PHP
	makeIndexer("PHP", "lsif-php", "github.com/davidrjenni/lsif-php", "davidrjenni/lsif-php"),

	// Python
	makeInternalIndexer("Python", "scip-python"),

	// Ruby
	makeInternalIndexer("Ruby", "scip-ruby"),

	// Rust
	makeInternalIndexer("Rust", "scip-rust"),
	makeIndexer("Rust", "rust-analyzer", "github.com/rust-lang/rust-analyzer"),

	// Terraform
	makeIndexer("Terraform", "lsif-terraform", "github.com/juliosueiras/lsif-terraform"),

	// TypeScript/JavaScript
	makeInternalIndexer("TypeScript", "scip-typescript"),
	makeInternalIndexer("TypeScript", "lsif-node"),
}

func makeInternalIndexer(key, name string) CodeIntelIndexer {
	return makeIndexer(
		key,
		name,
		fmt.Sprintf("github.com/sourcegraph/%s", name),
		fmt.Sprintf("sourcegraph/%s", name),
	)
}

func makeIndexer(key, name, urn string, dockerImages ...string) CodeIntelIndexer {
	return CodeIntelIndexer{
		LanguageKey:  key,
		Name:         name,
		URN:          urn,
		DockerImages: dockerImages,
	}
}

var ImageToIndexer = func() map[string]CodeIntelIndexer {
	m := map[string]CodeIntelIndexer{}
	for _, indexer := range AllIndexers {
		for _, dockerImage := range indexer.DockerImages {
			m[dockerImage] = indexer
		}
	}

	return m
}()

var PreferredIndexers = func() map[string]CodeIntelIndexer {
	preferred := map[string]CodeIntelIndexer{}

	m := map[string]CodeIntelIndexer{}
	for _, indexer := range AllIndexers {
		if p, ok := preferred[indexer.LanguageKey]; ok {
			m[indexer.Name] = p
		} else {
			m[indexer.Name] = indexer
			preferred[indexer.LanguageKey] = indexer
		}
	}

	return m
}()

var extensions = map[string][]string{
	"C++":        {".c", ".cp", ".cpp", ".cxx", ".h", ".hpp"},
	"Dart":       {".dart"},
	"DotNet":     {".cs", ".fs"},
	"Go":         {".go"},
	"HIE":        {".hs"},
	"Jsonnet":    {".jsonnet"},
	"JVM":        {".java", ".kt", ".scala"},
	"OCaml":      {".ml"},
	"PHP":        {".php"},
	"Python":     {".py"},
	"Ruby":       {".rb"},
	"Rust":       {".rs"},
	"Terraform":  {".tf"},
	"TypeScript": {".js", ".jsx", ".ts", ".tsx"},
}

// A map of file extension to a list of indexers in order of recommendation from most to least.
var LanguageToIndexer = func() map[string][]CodeIntelIndexer {
	m := map[string][]CodeIntelIndexer{}
	for _, indexer := range AllIndexers {
		for _, extension := range extensions[indexer.LanguageKey] {
			m[extension] = append(m[extension], indexer)
		}
	}

	return m
}()
