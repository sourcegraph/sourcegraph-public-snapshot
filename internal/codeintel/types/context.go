package types

import (
	codenavshared "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
)

type PreciseContext struct {
	Symbol            PreciseSymbolReference
	DefinitionSnippet string
	RepositoryName    string
	Filepath          string // TODO: redundant
	Location          codenavshared.UploadLocation
}

type PreciseSymbolReference struct {
	ScipName         string
	DescriptorSuffix string
	FuzzyName        *string
}
