import { escapeRegExp } from 'lodash'

/** The options that describe a search */
export interface SearchOptions {
    /** The query entered by the user */
    query: string
}

/**
 * Builds a URL query for given SearchOptions (without leading `?`)
 */
export function buildSearchURLQuery(options: SearchOptions): string {
    const searchParams = new URLSearchParams()
    searchParams.set('q', options.query)
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
    return {
        query: ['sq', 'q']
            .map(param => searchParams.get(param))
            .filter(s => s)
            .join(' '),
    }
}

export function searchQueryForRepoRev(repoPath: string, rev?: string): string {
    return `repo:${quoteIfNeeded(`^${escapeRegExp(repoPath)}$${rev ? `@${abbreviateOID(rev)}` : ''}`)} `
}

function abbreviateOID(oid: string): string {
    if (oid.length === 40) {
        return oid.slice(0, 7)
    }
    return oid
}

export function quoteIfNeeded(s: string): string {
    if (/["' ]/.test(s)) {
        return JSON.stringify(s)
    }
    return s
}
