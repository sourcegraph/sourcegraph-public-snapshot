import * as path from 'path'

import { cStyleComment } from './comments'
import { type FilterContext, type LanguageSpec, type Result, LSIFSupport } from './language-spec'
import { extractFromLines, filterResultsByImports, removeExtension } from './util'

/**
 * Filter a list of candidate definitions to select those likely to be valid
 * cross-references for a definition in this file. Accept candidates whose
 * path matches a relative import.
 *
 * If no candidates match, fall back to the raw (unfiltered) results so that
 * the user doesn't get an empty response unless there really is nothing.
 */
function filterDefinitions<T extends Result>(results: T[], { filePath, fileContent }: FilterContext): T[] {
    const importPaths = extractFromLines(fileContent, /\bfrom ["'](.*)["'];?$/, /\brequire\(["'](.*)["']\)/)

    return filterResultsByImports(
        results,
        importPaths,
        // Match results with a basename suffix of an import candidate
        ({ file }, importPath) => candidates(filePath, importPath).includes(removeExtension(file))
    )
}

/**
 * Construct a list of candidate paths that serve the given import path
 * relative to the given source path.
 */
function candidates(sourcePath: string, importPath: string): string[] {
    return [path.join(path.dirname(sourcePath), importPath), path.join(path.dirname(sourcePath), importPath, 'index')]
}

export const typescriptSpec: LanguageSpec = {
    languageID: 'typescript',
    additionalLanguages: ['javascript'],
    stylized: 'TypeScript',
    fileExts: ['ts', 'tsx', 'js', 'jsx'],
    commentStyles: [cStyleComment],
    filterDefinitions,
    lsifSupport: LSIFSupport.Robust,
    textDocumentImplemenationSupport: true,
}
