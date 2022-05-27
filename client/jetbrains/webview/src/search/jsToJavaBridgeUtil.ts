import { splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { ContentMatch } from '@sourcegraph/shared/src/search/stream'

import { loadContent } from './lib/blob'
import { PluginConfig, Theme, Search } from './types'

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

export async function getConfig(): Promise<PluginConfig> {
    try {
        return (await callJava({ action: 'getConfig' })) as PluginConfig
    } catch (error) {
        console.error(`Failed to get config: ${(error as Error).message}`)
        return {
            instanceURL: 'https://sourcegraph.com',
            isGlobbingEnabled: false,
            accessToken: null,
        }
    }
}

export async function getTheme(): Promise<Theme> {
    try {
        return (await callJava({ action: 'getTheme' })) as Theme
    } catch (error) {
        console.error(`Failed to get theme: ${(error as Error).message}`)
        return {
            isDarkTheme: true,
            buttonColor: '#0078d4',
        }
    }
}

export async function indicateFinishedLoading(): Promise<void> {
    try {
        await callJava({ action: 'indicateFinishedLoading' })
    } catch (error) {
        console.error(`Failed to indicate “finished loading”: ${(error as Error).message}`)
    }
}

export async function onPreviewChange(match: ContentMatch, lineMatchIndex: number): Promise<void> {
    const request = await createRequestForMatch(match, lineMatchIndex, 'preview')
    try {
        await callJava(request)
    } catch (error) {
        console.error(`Failed to preview match: ${(error as Error).message}`, request)
    }
}

export async function onPreviewClear(): Promise<void> {
    try {
        await callJava({ action: 'clearPreview' })
    } catch (error) {
        console.error(`Failed to clear preview: ${(error as Error).message}`)
    }
}

export async function onOpen(match: ContentMatch, lineMatchIndex: number): Promise<void> {
    try {
        await callJava(await createRequestForMatch(match, lineMatchIndex, 'open'))
    } catch (error) {
        console.error(`Failed to open match: ${(error as Error).message}`)
    }
}

export async function loadLastSearch(): Promise<Search | null> {
    try {
        return (await callJava({ action: 'loadLastSearch' })) as Search
    } catch (error) {
        console.error(`Failed to get last search: ${(error as Error).message}`)
        return null
    }
}

export function saveLastSearch(lastSearch: Search): void {
    callJava({ action: 'saveLastSearch', arguments: lastSearch })
        .then(() => {
            console.log(`Saved last search: ${JSON.stringify(lastSearch)}`)
        })
        .catch((error: Error) => {
            console.error(`Failed to save last search: ${error.message}`)
        })
}

async function callJava(request: Request): Promise<object> {
    return window.callJava(request)
}

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

// NOTE: This might be slow when the content is a really large file and the match is in the beginning of the file
// because we convert all rows to an array first.
// If we ever run into issues with large files, this is a place to get some wins.
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
