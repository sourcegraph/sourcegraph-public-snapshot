import type { Filter } from '$lib/shared'

interface FilterGroups {
    repo: Filter[]
    file: Filter[]
    lang: Filter[]
}

export function groupFilters(filters: Filter[] | null | undefined): FilterGroups {
    const groupedFilters: FilterGroups = {
        file: [],
        repo: [],
        lang: [],
    }
    if (filters) {
        for (const filter of filters) {
            switch (filter.kind) {
                case 'repo':
                case 'file':
                case 'lang': {
                    groupedFilters[filter.kind].push(filter)
                    break
                }
            }
        }
    }
    return groupedFilters
}
