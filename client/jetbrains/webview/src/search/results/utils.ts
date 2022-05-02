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

export function getIdForLine(match: ContentMatch, line: ContentMatch['lineMatches'][0]): string {
    return `${getIdForMatch(match)}-#-${match.lineMatches.indexOf(line)}`
}

export function decodeLineId(id: string): [matchId: string, lineNumber: number] {
    const [matchId, lineId] = id.split('-#-')
    return [matchId, parseInt(lineId, 10)]
}

export function getElementFromId(id: string): null | Element {
    // eslint-disable-next-line unicorn/prefer-query-selector
    return document.getElementById(`search-result-list-item-${id}`)
}

export function getNextResult(currentElement: Element): null | string {
    const next = currentElement.nextElementSibling
    if (next) {
        if (next.id) {
            return next.id.replace('search-result-list-item-', '')
        }
        return getNextResult(next)
    }
    return null
}

export function getPreviousResult(currentElement: Element): null | string {
    const previous = currentElement.previousElementSibling
    if (previous) {
        if (previous.id) {
            return previous.id.replace('search-result-list-item-', '')
        }
        return getPreviousResult(previous)
    }
    return null
}
