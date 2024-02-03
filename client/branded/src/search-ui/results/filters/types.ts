export enum FilterKind {
    Type = 'type',
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

export enum SearchTypeLabel {
    Code = 'Code',
    Repositories = 'Repositories',
    Paths = 'Paths',
    Symbols = 'Symbols',
    Commits = 'Commits',
    Diffs = 'Diffs',
}

export const SEARCH_TYPES_TO_FILTER_TYPES: Record<`${SearchTypeLabel}`, `${FilterKind}`[]> = {
    [SearchTypeLabel.Code]: [
        FilterKind.Language,
        FilterKind.Repository,
        FilterKind.File,
        FilterKind.Utility,
        FilterKind.Count,
    ],
    [SearchTypeLabel.Repositories]: [FilterKind.Repository, FilterKind.Utility, FilterKind.Count],
    [SearchTypeLabel.Paths]: [
        FilterKind.Language,
        FilterKind.Repository,
        FilterKind.File,
        FilterKind.Utility,
        FilterKind.Count,
    ],
    [SearchTypeLabel.Symbols]: [
        FilterKind.SymbolKind,
        FilterKind.Language,
        FilterKind.Repository,
        FilterKind.File,
        FilterKind.Utility,
        FilterKind.Count,
    ],
    [SearchTypeLabel.Commits]: [
        FilterKind.Author,
        FilterKind.Repository,
        FilterKind.CommitDate,
        FilterKind.Utility,
        FilterKind.Count,
    ],
    [SearchTypeLabel.Diffs]: [
        FilterKind.Author,
        FilterKind.Repository,
        FilterKind.CommitDate,
        FilterKind.Utility,
        FilterKind.Count,
    ],
}
