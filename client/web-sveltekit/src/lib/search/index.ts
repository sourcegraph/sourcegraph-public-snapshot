import { parseSearchURL, type ParsedSearchURL } from '$lib/web'

export interface ExtendedParsedSearchURL extends ParsedSearchURL {
    filters: string
    /**
     * Original query + filters.
     */
    filteredQuery: string | undefined
}

export function parseExtendedSearchURL(search: string): ExtendedParsedSearchURL {
    const parsedQuery = parseSearchURL(search)
    const filters = parseSearchFilters(search)
    const filteredQuery = filters ? parsedQuery.query + ' ' + filters : parsedQuery.query
    return { ...parsedQuery, filters, filteredQuery }
}

function parseSearchFilters(search: string): string {
    const params = new URLSearchParams(search)
    return params.get('filters') ?? ''
}
