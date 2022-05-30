import { ContentMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'

export function getFirstResultId(results: SearchMatch[]): string | null {
    const firstContentMatch: null | ContentMatch = results.find(result => result.type === 'content') as ContentMatch
    if (firstContentMatch) {
        return getResultId(firstContentMatch, firstContentMatch.lineMatches[0])
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
    return resultId.split('-#-')[0]
}

export function getResultId(
    match: SearchMatch,
    lineMatchOrSymbolName?: ContentMatch['lineMatches'][0] | string
): string {
    if (match.type === 'content') {
        return `${getMatchId(match)}-#-${match.lineMatches.indexOf(
            lineMatchOrSymbolName as ContentMatch['lineMatches'][0]
        )}`
    }

    if (match.type === 'symbol') {
        return `${getMatchId(match)}-#-${lineMatchOrSymbolName as string}`
    }

    return getMatchId(match)
}

export function getLineMatchIndexOrSymbolIndexForFileResult(resultId: string): number {
    return parseInt(resultId.split('-#-')[1], 10)
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
