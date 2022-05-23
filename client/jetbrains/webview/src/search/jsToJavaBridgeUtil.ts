import { splitPath } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ContentMatch } from '@sourcegraph/shared/src/search/stream'

import { Search } from './App'
import { loadContent } from './lib/blob'

interface MatchRequest {
    action: 'preview' | 'open'
    arguments: {
        fileName: string
        path: string
        content: string
        lineNumber: number
        absoluteOffsetAndLengths: number[][]
    }
}

interface GetConfigRequest {
    action: 'getConfig'
}

interface GetThemeRequest {
    action: 'getTheme'
}

interface SaveLastSearchRequest {
    action: 'saveLastSearch'
    arguments: Search
}

interface LoadLastSearchRequest {
    action: 'loadLastSearch'
}

interface ClearPreviewRequest {
    action: 'clearPreview'
}

interface IndicateFinishedLoadingRequest {
    action: 'indicateFinishedLoading'
}

export type Request =
    | MatchRequest
    | GetConfigRequest
    | GetThemeRequest
    | SaveLastSearchRequest
    | LoadLastSearchRequest
    | ClearPreviewRequest
    | IndicateFinishedLoadingRequest

export async function createRequestForMatch(
    match: ContentMatch,
    lineMatchIndex: number,
    action: MatchRequest['action']
): Promise<MatchRequest> {
    const fileName = splitPath(match.path)[1]
    const content = await loadContent(match)
    const characterCountUntilLine = getCharacterCountUntilLine(content, match.lineMatches[lineMatchIndex].lineNumber)
    const absoluteOffsetAndLengths = getAbsoluteOffsetAndLengths(
        match.lineMatches[lineMatchIndex].offsetAndLengths,
        characterCountUntilLine
    )

    return {
        action,
        arguments: {
            fileName,
            path: match.path,
            content,
            lineNumber: match.lineMatches[lineMatchIndex].lineNumber,
            absoluteOffsetAndLengths,
        },
    }
}

function getCharacterCountUntilLine(content: string, lineNumber: number): number {
    let count = 0
    const lines = content.split('\n') // This logic should handle \r\n well, too.
    for (let index = 0; index < lineNumber; index++) {
        count += lines[index].length + 1
    }
    console.log(`getCharacterCountUntilLine: ${count}`)
    return count
}

function getAbsoluteOffsetAndLengths(offsetAndLengths: number[][], characterCountUntilLine: number): number[][] {
    return offsetAndLengths.map(offsetAndLength => [offsetAndLength[0] + characterCountUntilLine, offsetAndLength[1]])
}
