import path from 'path'

import * as vscode from 'vscode'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'

import { JaccardMatch, bestJaccardMatch } from './bestJaccardMatch'
import { History } from './history'

/**
 * Keep property names in sync with the `EmbeddingsSearchResult` type.
 */
export interface ReferenceSnippet {
    fileName: string
    content: string
}

interface JaccardMatchWithFilename extends JaccardMatch {
    fileName: string
}

interface GetContextOptions {
    currentEditor: vscode.TextEditor
    history: History
    targetText: string
    windowSize: number
    maxChars: number
    codebaseContext: CodebaseContext
}

export async function getContext(options: GetContextOptions): Promise<ReferenceSnippet[]> {
    // TODO(valery): use codebaseContext here.
    const { currentEditor, history, targetText, windowSize, maxChars } = options
    const files = await getRelevantFiles(currentEditor, history)

    const matches: JaccardMatchWithFilename[] = []
    for (const { uri, contents } of files) {
        const match = bestJaccardMatch(targetText, contents, windowSize)
        if (!match) {
            continue
        }

        matches.push({
            // Use relative path to remove redundant information from the prompts and
            // keep in sync with embeddings search resutls which use relatve to repo root paths.
            fileName: path.normalize(vscode.workspace.asRelativePath(uri.fsPath)),
            ...match,
        })
    }

    matches.sort((a, b) => b.score - a.score)

    const context: ReferenceSnippet[] = []
    let totalChars = 0
    for (const match of matches) {
        if (totalChars + match.content.length > maxChars) {
            break
        }
        context.push(match)
        totalChars += match.content.length
    }

    return context
}

interface FileContents {
    uri: vscode.Uri
    contents: string
}

/**
 * Loads all relevant files for for a given text editor. Relevant files are defined as:
 *
 * - All currently open tabs matching the same language
 * - The last 10 files that were edited matching the same language
 *
 * For every file, we will load up to 10.000 lines to avoid OOMing when working with very large
 * files.
 */
async function getRelevantFiles(currentEditor: vscode.TextEditor, history: History): Promise<FileContents[]> {
    const files: FileContents[] = []

    const curLang = currentEditor.document.languageId
    if (!curLang) {
        return []
    }

    function addDocument(document: vscode.TextDocument): void {
        if (document.uri === currentEditor.document.uri) {
            // omit current file
            return
        }
        if (document.languageId !== curLang) {
            // TODO(beyang): handle JavaScript <-> TypeScript and verify this works for C header files
            // omit files of other languages
            return
        }

        // TODO(philipp-spiess): Find out if we have a better approach to truncate very large files.
        const endLine = Math.min(document.lineCount, 10_000)
        const range = new vscode.Range(0, 0, endLine, 0)

        files.push({
            uri: document.uri,
            contents: document.getText(range),
        })
    }

    const documents = vscode.workspace.textDocuments
    for (const document of documents) {
        if (document.fileName.endsWith('.git')) {
            // The VS Code API returns fils with the .git suffix for every open file
            continue
        }
        addDocument(document)
    }

    await Promise.all(
        history.lastN(10, curLang, [currentEditor.document.uri, ...files.map(f => f.uri)]).map(async item => {
            try {
                const document = await vscode.workspace.openTextDocument(item.document.uri)
                addDocument(document)
            } catch (error) {
                console.error(error)
            }
        })
    )
    return files
}
