import { splitPath } from '@sourcegraph/shared/src/components/RepoFileLink'
import { ContentMatch } from '@sourcegraph/shared/src/search/stream'

import { loadContent } from './lib/blob'

type RequestToJavaAction =
    | 'getConfig'
    | 'getTheme'
    | 'saveLastSearch'
    | 'loadLastSearch'
    | 'preview'
    | 'clearPreview'
    | 'open'

export interface RequestToJava {
    action: RequestToJavaAction
    arguments: object
}

export async function createRequestForMatch(
    match: ContentMatch,
    lineMatchIndex: number,
    action: RequestToJavaAction
): Promise<RequestToJava> {
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
