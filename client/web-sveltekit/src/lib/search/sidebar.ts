import {
    mdiCodeTags,
    mdiFileCodeOutline,
    mdiPlusMinus,
    mdiShapeOutline,
    mdiSourceCommit,
    mdiSourceRepository,
} from '@mdi/js'

import { FilterKind, FilterType, appendFilter, findFilter, updateFilter } from '$lib/shared'

function addOrUpdateTypeFilter(value: string): (query: string) => string {
    return query => {
        try {
            return updateFilter(query, FilterType.type, value)
        } catch {
            // Ignore error
        }
        return appendFilter(query, FilterType.type, value)
    }
}

function isSelectedTypeFilter(value: string): (query: string) => boolean {
    return query => findFilter(query, FilterType.type, FilterKind.Global)?.value?.value === value
}

interface ResultTypeFilter {
    label: string
    icon: string
    getQuery: (query: string) => string
    isSelected: (query: string) => boolean
}

export const resultTypeFilter: ResultTypeFilter[] = [
    {
        label: 'Code',
        icon: mdiCodeTags,
        getQuery: addOrUpdateTypeFilter('file'),
        isSelected: isSelectedTypeFilter('file'),
    },
    {
        label: 'Repositories',
        icon: mdiSourceRepository,
        getQuery: addOrUpdateTypeFilter('repo'),
        isSelected: isSelectedTypeFilter('repo'),
    },
    {
        label: 'Paths',
        icon: mdiFileCodeOutline,
        getQuery: addOrUpdateTypeFilter('path'),
        isSelected: isSelectedTypeFilter('path'),
    },
    {
        label: 'Symbols',
        icon: mdiShapeOutline,
        getQuery: addOrUpdateTypeFilter('symbol'),
        isSelected: isSelectedTypeFilter('symbol'),
    },
    {
        label: 'Commits',
        icon: mdiSourceCommit,
        getQuery: addOrUpdateTypeFilter('commit'),
        isSelected: isSelectedTypeFilter('commit'),
    },
    {
        label: 'Diffs',
        icon: mdiPlusMinus,
        getQuery: addOrUpdateTypeFilter('diff'),
        isSelected: isSelectedTypeFilter('diff'),
    },
]
