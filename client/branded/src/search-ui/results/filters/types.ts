export enum SearchFilterType {
    Code = 'Code',
    Repositories = 'Repositories',
    Paths = 'Paths',
    Symbols = 'Symbols',
    Commits = 'Commits',
    Diffs = 'Diffs',
}

export enum FiltersType {
    SymbolKind = 'symbol type',
    Language = 'lang',
    Author = 'author',
    Repository = 'repo',
    CommitDate = 'commit date',
    File = 'file',
    Utility = 'utility',

    // Synthetic filter, lives only on the client
    Count = 'count',
}

export const SEARCH_TYPES_TO_FILTER_TYPES: Record<SearchFilterType, `${FiltersType}`[]> = {
    [SearchFilterType.Code]: [
        FiltersType.Language,
        FiltersType.Repository,
        FiltersType.File,
        FiltersType.Utility,
        FiltersType.Count,
    ],
    [SearchFilterType.Repositories]: [FiltersType.Repository, FiltersType.Utility, FiltersType.Count],
    [SearchFilterType.Paths]: [
        FiltersType.Language,
        FiltersType.Repository,
        FiltersType.File,
        FiltersType.Utility,
        FiltersType.Count,
    ],
    [SearchFilterType.Symbols]: [
        FiltersType.SymbolKind,
        FiltersType.Language,
        FiltersType.Repository,
        FiltersType.File,
        FiltersType.Utility,
        FiltersType.Count,
    ],
    [SearchFilterType.Commits]: [
        FiltersType.Author,
        FiltersType.Repository,
        FiltersType.CommitDate,
        FiltersType.Utility,
        FiltersType.Count,
    ],
    [SearchFilterType.Diffs]: [
        FiltersType.Author,
        FiltersType.Repository,
        FiltersType.CommitDate,
        FiltersType.Utility,
        FiltersType.Count,
    ],
}
