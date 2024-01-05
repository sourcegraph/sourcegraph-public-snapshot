import type { Filter } from '@sourcegraph/shared/src/search/stream'

export enum SearchFilterType {
    Code = 'Code',
    Repositories = 'Repositories',
    Paths = 'Paths',
    Symbols = 'Symbols',
    Commits = 'Commits',
    Diffs = 'Diffs',
}

/**
 * Backend doesn't support all possible kind of filters yet, in order
 * to extend it with got this client-based filter type.
 */
export interface DynamicClientFilter extends Filter {
    kind: Filter['kind'] | 'select' | 'after' | 'before' | 'author'
}

export const SYMBOL_KIND_FILTERS: DynamicClientFilter[] = [
    { kind: 'select', label: 'Function', count: 0, exhaustive: true, value: 'select:symbol.function' },
    { kind: 'select', label: 'Method', count: 0, exhaustive: true, value: 'select:symbol.method' },
    { kind: 'select', label: 'Module', count: 0, exhaustive: true, value: 'select:symbol.module' },
    { kind: 'select', label: 'Class', count: 0, exhaustive: true, value: 'select:symbol.class' },
    { kind: 'select', label: 'Enum', count: 0, exhaustive: true, value: 'select:symbol.enum' },
]

export const COMMIT_DATE_FILTERS: DynamicClientFilter[] = [
    { kind: 'after', label: 'Last 24 hours', count: 0, exhaustive: true, value: 'after:yesterday' },
    { kind: 'before', label: 'Last week', count: 0, exhaustive: true, value: 'before:"1 week ago"' },
    { kind: 'before', label: 'Last month', count: 0, exhaustive: true, value: 'before:"1 month ago"' },
]
