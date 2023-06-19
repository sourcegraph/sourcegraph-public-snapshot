import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'

import { getContextFromEmbeddings } from './context-embeddings'
import { getContextFromCurrentEditor } from './context-local'
import { History } from './history'

/**
 * Keep property names in sync with the `EmbeddingsSearchResult` type.
 */
export interface ReferenceSnippet {
    fileName: string
    content: string
}

interface GetContextOptions {
    currentEditor: vscode.TextEditor
    history: History
    prefix: string
    suffix: string
    jaccardDistanceWindowSize: number
    maxChars: number
    codebaseContext: CodebaseContext
    isEmbeddingsContextEnabled?: boolean
}

export async function getContext(options: GetContextOptions): Promise<ReferenceSnippet[]> {
    const { maxChars, isEmbeddingsContextEnabled } = options

    /**
     * The embeddings context is sync to retrieve to keep the completions latency minumal.
     * If it's not available in cache yet, we'll retrieve it in the background and cache it for future use.
     */
    const embeddingsMatches = isEmbeddingsContextEnabled ? getContextFromEmbeddings(options) : []
    const editorMatches = await getContextFromCurrentEditor(options)

    const usedFilenames = new Set<string>()
    const context: ReferenceSnippet[] = []
    let totalChars = 0

    /**
     * Iterate over matches and add them to the context.
     * Discard editor matches for files with embedding matches.
     */
    for (const match of [...embeddingsMatches, ...editorMatches]) {
        const existingMatch = usedFilenames.has(match.fileName)

        if (!existingMatch) {
            usedFilenames.add(match.fileName)

            if (totalChars + match.content.length > maxChars) {
                break
            }
            context.push(match)
            totalChars += match.content.length
        }
    }

    return context
}
