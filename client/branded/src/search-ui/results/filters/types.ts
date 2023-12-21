import type { Filter } from '@sourcegraph/shared/src/search/stream'

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

export const SYMBOL_KIND_FILTERS: Filter[] = [
    { kind: 'select', label: 'Function', count: 0, limitHit: false, value: 'select:symbol.function' },
    { kind: 'select', label: 'Method', count: 0, limitHit: false, value: 'select:symbol.method' },
    { kind: 'select', label: 'Module', count: 0, limitHit: false, value: 'select:symbol.module' },
    { kind: 'select', label: 'Class', count: 0, limitHit: false, value: 'select:symbol.class' },
    { kind: 'select', label: 'Enum', count: 0, limitHit: false, value: 'select:symbol.enum' },
]

export const COMMIT_DATE_FILTERS: Filter[] = [
    { kind: 'after', label: 'Last 24 hours', count: 0, limitHit: false, value: 'after:yesterday' },
    { kind: 'before', label: 'Last week', count: 0, limitHit: false, value: 'before:"1 week ago"' },
    { kind: 'before', label: 'Last month', count: 0, limitHit: false, value: 'before:"1 month ago"' },
]
