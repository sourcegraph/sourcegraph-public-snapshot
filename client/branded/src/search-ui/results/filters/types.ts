export enum SearchFilterType {
    Code = 'Code',
    Repositories = 'Repositories',
    Paths = 'Paths',
    Symbols = 'Symbols',
    Commits = 'Commits',
    Diffs = 'Diffs',
}

export enum SearchResultFilters {
    ByRepository,
    ByLanguage,
    ByPath,
    ByMetadata,
    Recipes,
    ArchivedAndForked,
    BySymbolKind,
    ByAuthor,
    ByCommitDate,
    ByDiffType,
}
