import type { LineMatch, SearchMatch, MatchedSymbol } from '@sourcegraph/shared/src/search/stream'

const SUPPORTED_TYPES = new Set(['commit', 'content', 'path', 'symbol', 'repo'])
const ID_SEPERATOR = '-#-'

export function getFirstResultId(results: SearchMatch[]): string | null {
    const firstSupportedMatch: null | SearchMatch = results.find(result => SUPPORTED_TYPES.has(result.type)) ?? null

    if (firstSupportedMatch) {
        return getResultId(
            firstSupportedMatch,
            firstSupportedMatch.type === 'content'
                ? firstSupportedMatch.lineMatches
                    ? firstSupportedMatch.lineMatches[0]
                    : undefined
                : firstSupportedMatch.type === 'symbol'
                ? firstSupportedMatch.symbols[0]
                : undefined
        )
    }
    return null
}

export function getMatchId(match: SearchMatch): string {
    if (match.type === 'commit') {
        return `${match.repository}-${match.oid.slice(0, 7)}`
    }

    if (match.type === 'content' || match.type === 'path' || match.type === 'symbol') {
        return `${match.repository}-${match.path}`
    }

    if (match.type === 'repo') {
        return match.repository
    }

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore This is here in preparation for future match types
    console.error('Unknown match type:', match.type)
    return ''
}

export function getMatchIdForResult(resultId: string): string {
    return resultId.split(ID_SEPERATOR)[0]
}

export type LineMatchItem = LineMatch
export type SymbolMatchItem = MatchedSymbol
export function getResultId(match: SearchMatch, lineOrSymbolMatch?: LineMatchItem | SymbolMatchItem): string {
    if (match.type === 'content') {
        return `${getMatchId(match)}${ID_SEPERATOR}${match.lineMatches?.indexOf(lineOrSymbolMatch as LineMatchItem)}`
    }
    if (match.type === 'symbol') {
        return `${getMatchId(match)}${ID_SEPERATOR}${match.symbols.indexOf(lineOrSymbolMatch as SymbolMatchItem)}`
    }
    return getMatchId(match)
}

export function getLineOrSymbolMatchIndexForFileResult(resultId: string): number {
    return parseInt(resultId.split(ID_SEPERATOR)[1], 10)
}

export function getSearchResultElement(resultId: string): null | Element {
    // eslint-disable-next-line unicorn/prefer-query-selector
    return document.getElementById(`search-result-list-item-${resultId}`)
}

export function getSiblingResultElement(currentElement: Element, direction: 'previous' | 'next'): null | Element {
    const sibling: Element | null =
        direction === 'previous' ? currentElement.previousElementSibling : currentElement.nextElementSibling
    if (sibling) {
        if (sibling.id) {
            return sibling
        }
        return getSiblingResultElement(sibling, direction)
    }
    return null
}
