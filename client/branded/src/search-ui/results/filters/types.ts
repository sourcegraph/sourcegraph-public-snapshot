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

export const TYPES_TO_FILTERS = {
    [SearchFilterType.Code]: [
        SearchResultFilters.ByLanguage,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ByPath,
        SearchResultFilters.Recipes,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Repositories]: [
        SearchResultFilters.ByLanguage,
        SearchResultFilters.ByMetadata,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Paths]: [
        SearchResultFilters.ByLanguage,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Symbols]: [
        SearchResultFilters.BySymbolKind,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ByPath,
    ],
    [SearchFilterType.Commits]: [
        SearchResultFilters.ByAuthor,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ByCommitDate,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Diffs]: [
        SearchResultFilters.ByDiffType,
        SearchResultFilters.ByAuthor,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ArchivedAndForked,
    ],
}
