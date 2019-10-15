import { escapeRegExp } from 'lodash'
import { SearchPatternType } from '../../../shared/src/graphql/schema'

/**
 * Parses the query out of the URL search params (the 'q' parameter). If the 'q' parameter is not present, it
 * returns undefined.
 */
export function parseSearchURLQuery(query: string): string | undefined {
    const searchParams = new URLSearchParams(query)
    return searchParams.get('q') || undefined
}

/**
 * Parses the pattern type out of the URL search params (the 'patternType' parameter). If the 'pattern' parameter
 * is not present, or it is an invalid value, it returns undefined.
 */
export function parseSearchURLPatternType(query: string): SearchPatternType | undefined {
    const searchParams = new URLSearchParams(query)
    const patternType = searchParams.get('patternType')
    if (patternType !== SearchPatternType.literal && patternType !== SearchPatternType.regexp) {
        return undefined
    }
    return patternType
}

export function searchQueryForRepoRev(repoName: string, rev?: string): string {
    return `repo:${quoteIfNeeded(`^${escapeRegExp(repoName)}$${rev ? `@${abbreviateOID(rev)}` : ''}`)} `
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

export interface PatternTypeProps {
    patternType: SearchPatternType
    togglePatternType: () => void
}
