import * as path from 'path'

import { pythonStyleComment } from './comments'
import type { FilterContext, LanguageSpec, Result } from './language-spec'
import { extractFromLines, filterResultsByImports, removeExtension } from './util'

/**
 * Filter a list of candidate definitions to select those likely to be valid
 * cross-references for a definition in this file. Accept candidates whose
 * path matches an absolute or relative (to the current file) import.
 *
 * If no candidates match, fall back to the raw (unfiltered) results so that
 * the user doesn't get an empty response unless there really is nothing.
 */
function filterDefinitions<T extends Result>(results: T[], { filePath, fileContent }: FilterContext): T[] {
    const importPaths = extractFromLines(fileContent, /^import ([\w.]*)/, /^from ([\w.]*)/)

    return filterResultsByImports(results, importPaths, ({ file }, importPath) => {
        const relativePath = relativeImportPath(filePath, importPath)
        if (relativePath) {
            // Match results imported relatively
            return relativePath === removeExtension(file)
        }

        // Match results imported absolutely
        return file.includes(absoluteImportPath(importPath))
    })
}

/**
 * Converts an absolute Python import path into a file path.
 *
 * @param importPath The absolute Python import path.
 */
function absoluteImportPath(importPath: string): string {
    return importPath.replaceAll('.', '/')
}

/**
 * Converts a Python import path into a file path relative to the
 * given source path. If the import path is not relative, method
 * function returns undefined.
 *
 * @param sourcePath The source file path relative to the repository root.
 * @param importPath The relative or absolute Python import path (`.a.b`, `a.b.c`).
 */
export function relativeImportPath(sourcePath: string, importPath: string): string | undefined {
    const match = /^\.(\.*)(.*)/.exec(importPath)
    if (!match) {
        return undefined
    }
    const [, parentDots, rest] = match

    return path.join(path.dirname(sourcePath), '../'.repeat(parentDots.length), rest.replaceAll('.', '/'))
}

export const pythonSpec: LanguageSpec = {
    languageID: 'python',
    stylized: 'Python',
    fileExts: ['py'],
    commentStyles: [pythonStyleComment],
    filterDefinitions,
}
