import path from 'path'

import { differenceInMinutes } from 'date-fns'
import { LRUCache } from 'lru-cache'
import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'

import type { ReferenceSnippet } from './context'
import { logCompletionEvent } from './logger'

interface Options {
    currentEditor: vscode.TextEditor
    prefix: string
    suffix: string
    codebaseContext: CodebaseContext
}

interface EmbeddingsForFile {
    embeddings: ReferenceSnippet[]
    lastChange: Date
}

const embeddingsPerFile = new LRUCache<string, EmbeddingsForFile>({
    max: 10,
})

export function getContextFromEmbeddings(options: Options): ReferenceSnippet[] {
    const { currentEditor, codebaseContext, prefix, suffix } = options

    const currentFilePath = path.normalize(currentEditor.document.fileName)
    const embeddingsForCurrentFile = embeddingsPerFile.get(currentFilePath)

    /**
     * Fetch embeddings if we don't have any or if the last fetch was more than 5 minutes ago.
     * Ideally, we should fetch embeddings in the background if file significantly changed.
     * We can use the `onDidChangeTextDocument` event with some diffing logic for that in the improved version.
     */
    if (!embeddingsForCurrentFile || differenceInMinutes(embeddingsForCurrentFile.lastChange, new Date()) > 5) {
        fetchAndSaveEmbeddings({
            codebaseContext,
            currentFilePath,
            // Use preifx + suffix to limit number of lines we send to the server.
            // We can use the fullText here via `currentEditor.document.getText()` but
            // it can negatively affect the embeddings quality and price.
            text: prefix + suffix,
        }).catch(console.error)
    }

    // Return embeddings for current file if we have any in the cache.
    return embeddingsForCurrentFile?.embeddings || []
}

interface FetchEmbeddingsOptions {
    currentFilePath: string
    text: string
    codebaseContext: CodebaseContext
}

const NUM_CODE_RESULTS = 2
// Query a bigger number of code results than we need to get embeddings NOT from the current file.
const NUM_CODE_RESULTS_EXTRA = 5
const NUM_TEXT_RESULTS = 1

async function fetchAndSaveEmbeddings(options: FetchEmbeddingsOptions): Promise<void> {
    const { currentFilePath, text, codebaseContext } = options

    logCompletionEvent('fetchEmbeddings')

    // TODO: add comment on how big are the embedding results
    // TODO: add comment on what's the price for embedding a file to run embeddings search
    const { results } = await codebaseContext.getSearchResults(text, {
        numCodeResults: NUM_CODE_RESULTS + NUM_CODE_RESULTS_EXTRA,
        numTextResults: NUM_TEXT_RESULTS,
    })

    const embeddingResultsWithoutCurrentFile = results
        .map(result => ({
            ...result,
            fileName: path.normalize(result.fileName),
        }))
        .filter(result => !currentFilePath.endsWith(result.fileName))

    embeddingsPerFile.set(currentFilePath, {
        embeddings: embeddingResultsWithoutCurrentFile.slice(0, NUM_CODE_RESULTS + NUM_TEXT_RESULTS - 1),
        lastChange: new Date(),
    })
}
