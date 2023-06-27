import path from 'path'

import { History, CompletionsTextEditor, LightTextDocument } from '.'
import { JaccardMatch, bestJaccardMatch } from './bestJaccardMatch'
import type { ReferenceSnippet } from './context'

interface JaccardMatchWithFilename extends JaccardMatch {
    fileName: string
}

interface Options {
    currentEditor: CompletionsTextEditor
    history: History
    prefix: string
    jaccardDistanceWindowSize: number
}

export async function getContextFromCurrentEditor(options: Options): Promise<ReferenceSnippet[]> {
    const { currentEditor, history, prefix, jaccardDistanceWindowSize } = options

    const targetText = lastNLines(prefix, jaccardDistanceWindowSize)
    const files = await getRelevantFiles(currentEditor, history)

    const matches: JaccardMatchWithFilename[] = []
    for (const { uri, contents } of files) {
        const match = bestJaccardMatch(targetText, contents, jaccardDistanceWindowSize)
        if (!match) {
            continue
        }

        matches.push({
            // Use relative path to remove redundant information from the prompts and
            // keep in sync with embeddings search resutls which use relatve to repo root paths.
            fileName: path.normalize((await currentEditor.getDocumentRelativePath(uri))!),
            ...match,
        })
    }

    matches.sort((a, b) => b.score - a.score)

    return matches
}

interface FileContents {
    uri: string
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
async function getRelevantFiles(currentEditor: CompletionsTextEditor, history: History): Promise<FileContents[]> {
    const files: FileContents[] = []

    const curLang = currentEditor.getCurrentDocument()!.languageId
    if (!curLang) {
        return []
    }

    async function addDocument(document: LightTextDocument): Promise<void> {
        if (document.uri === currentEditor.getCurrentDocument()!.uri) {
            // omit current file
            return
        }
        if (document.languageId !== curLang) {
            // TODO(beyang): handle JavaScript <-> TypeScript and verify this works for C header files
            // omit files of other languages
            return
        }

        // TODO(philipp-spiess): Find out if we have a better approach to truncate very large files.
        files.push({
            uri: document.uri,
            contents: (await currentEditor.getDocumentTextTruncated(document.uri))!,
        })
    }

    for (const document of currentEditor.getOpenDocuments()) {
        if (document.uri.endsWith('.git')) {
            // The VS Code API returns fils with the .git suffix for every open file
            continue
        }
        await addDocument(document)
    }

    await Promise.all(
        history
            .lastN(10, curLang, [currentEditor.getCurrentDocument()!.uri, ...files.map(f => f.uri)])
            .map(async item => {
                try {
                    await addDocument(item)
                } catch (error) {
                    console.error(error)
                }
            })
    )
    return files
}

function lastNLines(text: string, n: number): string {
    const lines = text.split('\n')
    return lines.slice(Math.max(0, lines.length - n)).join('\n')
}
