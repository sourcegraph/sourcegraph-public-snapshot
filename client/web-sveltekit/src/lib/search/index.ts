import { KeywordKind, scanSearchQuery } from '$lib/shared'
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

    let filteredQuery = parsedQuery.query

    if (filteredQuery && filters) {
        // We need to wrap the query in parenthesis if it contains AND or OR operators to avoid
        // precedence issues.
        const result = scanSearchQuery(filteredQuery)
        if (result.type === 'success') {
            // Only append filters if the main query is valid
            if (
                result.term.some(
                    token =>
                        token.type === 'keyword' && (token.kind === KeywordKind.And || token.kind === KeywordKind.Or)
                )
            ) {
                filteredQuery = `(${filteredQuery})`
            }
            filteredQuery += ' ' + filters
        }
    }

    return { ...parsedQuery, filters, filteredQuery }
}

function parseSearchFilters(search: string): string {
    const params = new URLSearchParams(search)
    return params.get('filters') ?? ''
}
