package types

type PreciseContext struct {
	Symbol            PreciseSymbolReference
	DefinitionSnippet string
	RepositoryName    string
	Filepath          string
}

type PreciseSymbolReference struct {
	ScipName         string
	DescriptorSuffix string
	FuzzyName        *string
}
