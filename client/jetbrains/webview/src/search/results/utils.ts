import { ContentMatch, SearchMatch } from '@sourcegraph/shared/src/search/stream'

export function getFirstResultId(results: SearchMatch[]): string | null {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const firstContentMatch: null | ContentMatch = results.find(result => result.type === 'content')
    if (firstContentMatch) {
        return getIdForLine(firstContentMatch, firstContentMatch.lineMatches[0])
    }
    return null
}

export function getIdForMatch(match: ContentMatch): string {
    return `${match.repository}-${match.path}`
}

export function getIdForLine(match: ContentMatch, lineMatch: ContentMatch['lineMatches'][0]): string {
    return `${getIdForMatch(match)}-#-${match.lineMatches.indexOf(lineMatch)}`
}

export function decodeLineId(id: string): [matchId: string, lineMatchIndex: number] {
    const [matchId, lineMatchIndex] = id.split('-#-')
    return [matchId, parseInt(lineMatchIndex, 10)]
}

export function getElementFromId(id: string): null | Element {
    // eslint-disable-next-line unicorn/prefer-query-selector
    return document.getElementById(`search-result-list-item-${id}`)
}

export function getSiblingResult(currentElement: Element, direction: 'previous' | 'next'): null | string {
    const sibling = direction === 'previous' ? currentElement.previousElementSibling : currentElement.nextElementSibling
    if (sibling) {
        if (sibling.id) {
            return sibling.id.replace('search-result-list-item-', '')
        }
        return getSiblingResult(sibling, direction)
    }
    return null
}
