import {
    mdiCodeTags,
    mdiFileCodeOutline,
    mdiPlusMinus,
    mdiShapeOutline,
    mdiSourceCommit,
    mdiSourceRepository,
} from '@mdi/js'

import type { Filter } from '@sourcegraph/shared/src/search/stream'

export type URLQueryFilter = Pick<Filter, 'kind' | 'label' | 'value'>

export interface SidebarFilter extends Filter {
    selected: boolean
}

const DYNAMIC_FILTER_URL_QUERY_KEY = 'df'

function deserializeURLFilter(input: string): URLQueryFilter {
    const [kind, label, value] = JSON.parse(input) as [Filter['kind'], string, string]
    return { kind, label, value }
}

function serializeURLFilter(input: URLQueryFilter): string {
    return JSON.stringify([input.kind, input.label, input.value])
}

export function updateFilterInURL(url: URL, filter: URLQueryFilter, remove: boolean): URL {
    let selectedFilters = filtersFromURL(url)
    if (remove) {
        selectedFilters = selectedFilters.filter(
            selectedFilter => selectedFilter.kind !== filter.kind || selectedFilter.value != filter.value
        )
    } else {
        selectedFilters.push(filter)
    }

    const newURL = new URL(url)
    newURL.searchParams.delete(DYNAMIC_FILTER_URL_QUERY_KEY)
    selectedFilters
        .map(serializeURLFilter)
        .forEach(selectedFilter => newURL.searchParams.append(DYNAMIC_FILTER_URL_QUERY_KEY, selectedFilter))

    return newURL
}

export function filtersFromURL(url: URL): URLQueryFilter[] {
    return url.searchParams.getAll(DYNAMIC_FILTER_URL_QUERY_KEY).map(deserializeURLFilter)
}

export const staticTypeFilters: URLQueryFilter[] = [
    { kind: 'type', label: 'Code', value: 'type:file' },
    { kind: 'type', label: 'Repositories', value: 'type:repo' },
    { kind: 'type', label: 'Paths', value: 'type:path' },
    { kind: 'type', label: 'Symbols', value: 'type:symbol' },
    { kind: 'type', label: 'Commits', value: 'type:commit' },
    { kind: 'type', label: 'Diffs', value: 'type:diff' },
]

export const typeFilterIcons = new Map([
    ['Code', mdiCodeTags],
    ['Repositories', mdiSourceRepository],
    ['Paths', mdiFileCodeOutline],
    ['Symbols', mdiShapeOutline],
    ['Commits', mdiSourceCommit],
    ['Diffs', mdiPlusMinus],
])

export type FilterGroups = Record<Filter['kind'], SidebarFilter[]>

export function groupFilters(filters: Filter[], selectedFilters: URLQueryFilter[]): FilterGroups {
    const groupedFilters: FilterGroups = {
        type: [],
        repo: [],
        lang: [],
        utility: [],
        author: [],
        'commit date': [],
        'symbol type': [],
    }
    for (const filter of filters) {
        const selected = selectedFilters.some(
            selectedFilter => selectedFilter.kind === filter.kind && selectedFilter.value === filter.value
        )
        groupedFilters[filter.kind].push({ ...filter, selected })
    }
    return groupedFilters
}
