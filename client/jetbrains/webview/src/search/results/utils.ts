import { ContentMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'

export function getFirstResultId(results: SearchMatch[]): string | null {
    const firstContentMatch: null | ContentMatch = results.find(result => result.type === 'content') as ContentMatch
    if (firstContentMatch) {
        return getResultIdForContentMatch(firstContentMatch, firstContentMatch.lineMatches[0])
    }
    return null
}

export function getContentMatchId(match: ContentMatch): string {
    return `${match.repository}-${match.path}`
}

export function getResultIdForContentMatch(match: ContentMatch, lineMatch: ContentMatch['lineMatches'][0]): string {
    return `${getContentMatchId(match)}-#-${match.lineMatches.indexOf(lineMatch)}`
}

export function splitResultIdForContentMatch(resultId: string): [matchId: string, lineMatchIndex: number] {
    const [fileId, lineMatchIndex] = resultId.split('-#-')
    return [fileId, parseInt(lineMatchIndex, 10)]
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
