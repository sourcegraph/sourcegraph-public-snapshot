/** The options that describe a search */
export interface SearchOptions {
    /** The query entered by the user */
    query: string

    /**
     * The query provided by the active scope. If undefined,
     * the last-active scope should be used.
     */
    scopeQuery: string | undefined
}

/**
 * Builds a URL query for given SearchOptions (without leading `?`)
 */
export function buildSearchURLQuery(options: SearchOptions): string {
    const searchParams = new URLSearchParams()
    searchParams.set('q', options.query)
    if (typeof options.scopeQuery === 'string') {
        searchParams.set('sq', options.scopeQuery)
    }
    return searchParams
        .toString()
        .replace(/%2F/g, '/')
        .replace(/%3A/g, ':')
}

/**
 * Parses the SearchOptions out of URL search params. If neither the 'q' nor
 * 'sq' params are present, it returns undefined.
 */
export function parseSearchURLQuery(query: string): SearchOptions | undefined {
    const searchParams = new URLSearchParams(query)
    if (!searchParams.has('q') && !searchParams.has('sq')) {
        return undefined
    }
    const sq = searchParams.get('sq')
    return {
        query: searchParams.get('q') || '',
        scopeQuery: sq !== null ? sq : undefined,
    }
}

/**
 * Returns whether the two sets of search options are equal.
 */
export function searchOptionsEqual(a: SearchOptions, b: SearchOptions): boolean {
    return a.query === b.query && a.scopeQuery === b.scopeQuery
}

/**
 * Returns the URL without the search options URL query params ('q' and 'sq').
 */
export function urlWithoutSearchOptions(location: Location): string {
    const params = new URLSearchParams(location.search)
    params.delete('q')
    params.delete('sq')
    const query = Array.from(params.keys()).length > 0 ? `?${params.toString()}` : ''
    return location.protocol + '//' + location.host + location.pathname + query + location.hash
}
