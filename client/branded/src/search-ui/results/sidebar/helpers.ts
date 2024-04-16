import { useState, useEffect, useRef } from 'react'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter as QueryFilter } from '@sourcegraph/shared/src/search/query/token'
import type { Filter } from '@sourcegraph/shared/src/search/stream'

/**
 * Given a search query and filters from query results, this hook will return
 * a repo name if
 * - 'filters' is not empty and only contains a single repo filter
 * - 'filters' is empty, the query contains only a single repo filter and has the
 * same value as a previous non-empty search
 *
 * In all other cases it will return an empty string.
 */
export function useLastRepoName(query: string, filters: Filter[] = []): string {
    const lastRepoQuery = useRef('')
    const [repoName, setRepoName] = useState('')

    useEffect(() => {
        // Determine whether query contains a single repo filter and remember it
        // if it exists
        let queryRepoFilters: QueryFilter[] = []
        let queryRepo = ''
        const scanResult = scanSearchQuery(query)
        if (scanResult.type === 'success') {
            queryRepoFilters = findFilters(scanResult.term, FilterType.repo)
            if (queryRepoFilters.length === 1) {
                queryRepo = queryRepoFilters[0].value?.value ?? ''
            }
        }
        const repoFilters = getFiltersOfKind(filters, FilterType.repo)
        switch (repoFilters.length) {
            case 0: {
                // Reuse last repo name if query contains a repo filter and
                // it's the same as the previous one, otherwise clear previous
                // repo name
                if (!queryRepo || queryRepo !== lastRepoQuery.current) {
                    lastRepoQuery.current = ''
                    setRepoName('')
                }
                break
            }
            case 1: {
                // Update last repo name and repo query
                if (queryRepoFilters.length === 0) {
                    // This is a special case: If the query doesn't contain a
                    // repo but the results are from a single repo, we record the
                    // repo name from the search results. This way, if one of the
                    // rev links is clicked and we add both, rev and repo filters
                    // (see ./Revisions.tsx), the revisions section won't disappear.
                    queryRepo = `^${repoFilters[0].label}$`
                }
                lastRepoQuery.current = queryRepo
                setRepoName(repoFilters[0].label)
                break
            }
            default: {
                // multiple repos are matched, clear everything
                lastRepoQuery.current = ''
                setRepoName('')
            }
        }
    }, [query, filters, lastRepoQuery])

    return repoName
}

export function getFiltersOfKind(filters: Filter[] = [], kind: FilterType): Filter[] {
    return filters.filter(filter => filter.kind === kind && filter.value !== '')
}
