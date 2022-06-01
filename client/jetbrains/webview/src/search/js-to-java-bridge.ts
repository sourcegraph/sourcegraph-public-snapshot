import { encode } from 'js-base64'

import { splitPath } from '@sourcegraph/shared/src/components/RepoLink'
import { ContentMatch, PathMatch, SearchMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'

import { loadContent } from './lib/blob'
import { PluginConfig, Search, Theme } from './types'

export interface MatchRequest {
    action: 'preview' | 'open'
    arguments: {
        fileName: string
        path: string
        content: string | null
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

export interface SaveLastSearchRequest {
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

export async function onPreviewChange(match: SearchMatch, lineMatchIndexOrSymbolIndex?: number): Promise<void> {
    const request = await createPreviewOrOpenRequest(match, lineMatchIndexOrSymbolIndex, 'preview')
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

export async function onOpen(match: SearchMatch, lineMatchIndexOrSymbolIndex?: number): Promise<void> {
    const request = await createPreviewOrOpenRequest(match, lineMatchIndexOrSymbolIndex, 'open')
    if (request.arguments.fileName) {
        try {
            await callJava(request)
        } catch (error) {
            console.error(`Failed to open match: ${(error as Error).message}`)
        }
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

export async function createPreviewOrOpenRequest(
    match: SearchMatch,
    lineMatchIndexOrSymbolIndex: number | undefined,
    action: MatchRequest['action']
): Promise<MatchRequest> {
    if (match.type === 'commit') {
        return {
            action,
            arguments: {
                fileName: '',
                path: '',
                content: match.message,
                lineNumber: -1,
                absoluteOffsetAndLengths: [],
            },
        }
    }

    if (match.type === 'content') {
        return createPreviewOrOpenRequestForContentMatch(match, lineMatchIndexOrSymbolIndex as number, action)
    }

    if (match.type === 'path') {
        return createPreviewOrOpenRequestForPathMatch(match, action)
    }

    if (match.type === 'repo') {
        return {
            action,
            arguments: {
                fileName: '',
                path: '',
                content: null,
                lineNumber: -1,
                absoluteOffsetAndLengths: [],
            },
        }
    }

    if (match.type === 'symbol') {
        return createPreviewOrOpenRequestForSymbolMatch(match, lineMatchIndexOrSymbolIndex as number, action)
    }

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore This is here in preparation for future match types
    console.log(`Unknown match type: “${match.type}”`)

    return {
        action,
        arguments: {
            fileName: '',
            path: '',
            content: null,
            lineNumber: -1,
            absoluteOffsetAndLengths: [],
        },
    }
}

export async function createPreviewOrOpenRequestForContentMatch(
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
            content: prepareContent(content),
            lineNumber: match.lineMatches[lineMatchIndex].lineNumber,
            absoluteOffsetAndLengths,
        },
    }
}

export async function createPreviewOrOpenRequestForPathMatch(
    match: PathMatch,
    action: MatchRequest['action']
): Promise<MatchRequest> {
    const fileName = splitPath(match.path)[1]
    const content = await loadContent(match)

    return {
        action,
        arguments: {
            fileName,
            path: match.path,
            content: prepareContent(content),
            lineNumber: -1,
            absoluteOffsetAndLengths: [],
        },
    }
}

export async function createPreviewOrOpenRequestForSymbolMatch(
    match: SymbolMatch,
    symbolIndex: number,
    action: MatchRequest['action']
): Promise<MatchRequest> {
    const fileName = splitPath(match.path)[1]
    const content = await loadContent(match)

    return {
        action,
        arguments: {
            fileName,
            path: match.path,
            content: prepareContent(content),
            lineNumber: -1,
            absoluteOffsetAndLengths: [],
        },
    }
}

// We encode the content as base64-encoded string to avoid encoding errors in the Java JSON parser.
// The Java side also does not expact `\r\n` line endings, so we replace them with `\n`.
//
// We can not use the native btoa() function because it does not support all Unicode characters.
function prepareContent(content: string | null): string | null {
    if (content === null) {
        return null
    }
    return encode(content.replaceAll('\r\n', '\n'))
}

// NOTE: This might be slow when the content is a really large file and the match is in the
// beginning of the file because we convert all rows to an array first.
//
// If we ever run into issues with large files, this is a place to get some wins.
function getCharacterCountUntilLine(content: string | null, lineNumber: number): number {
    if (content === null) {
        return 0
    }

    let count = 0
    const lines = content.replaceAll('\r\n', '\n').split('\n')
    for (let index = 0; index < lineNumber; index++) {
        count += lines[index].length + 1
    }
    return count
}

function getAbsoluteOffsetAndLengths(offsetAndLengths: number[][], characterCountUntilLine: number): number[][] {
    return offsetAndLengths.map(offsetAndLength => [offsetAndLength[0] + characterCountUntilLine, offsetAndLength[1]])
}
