import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Keyword } from '@sourcegraph/shared/src/search/query/token'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { useSyncedWithURLState } from '@sourcegraph/wildcard'

export type URLQueryFilter = Pick<Filter, 'kind' | 'label' | 'value'>

export function serializeURLQueryFilters(filters: URLQueryFilter[]): string | null {
    if (filters.length === 0) {
        return null
    }
    return JSON.stringify(filters.map(f => [f.kind, f.label, f.value]))
}

export function deserializeURLQueryFilters(serialized: string | null): URLQueryFilter[] {
    if (!serialized) {
        return []
    }
    const parsed = JSON.parse(serialized) as [Filter['kind'], string, string][]
    return parsed.map(([kind, label, value]) => ({ kind, label, value }))
}

export function useUrlFilters(): [URLQueryFilter[], (newFilters: URLQueryFilter[]) => void] {
    const [filterQuery, setFilterQuery] = useSyncedWithURLState<URLQueryFilter[], string>({
        urlKey: 'filters',
        serializer: serializeURLQueryFilters,
        deserializer: deserializeURLQueryFilters,
        replace: false,
    })

    return [filterQuery, setFilterQuery]
}

export function mergeQueryAndFilters(query: string, filters: URLQueryFilter[]): string {
    const tokens = scanSearchQuery(query)

    // Return original query if it's non-valid query
    if (tokens.type === 'error') {
        return query
    }

    const filterQuery = filters.map(f => f.value).join(' ')
    const keywords = tokens.term.filter(token => token.type === 'keyword') as Keyword[]
    const hasAnd = keywords.some(filter => filter.kind === 'and')
    const hasOr = keywords.some(filter => filter.kind === 'or')

    // Wrap original query with parenthesize if the query has 'or' or 'and'
    // operators, otherwise simple concatenation may not work for all cases.
    if (hasOr || hasAnd) {
        return `(${query}) ${filterQuery}`.trim()
    }

    return `${query} ${filterQuery}`.trim()
}
