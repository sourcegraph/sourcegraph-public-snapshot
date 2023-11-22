import * as path from 'path'

import { slashPattern } from './comments'
import { type FilterContext, type LanguageSpec, type Result, LSIFSupport } from './language-spec'
import { extractFromLines, filterResults } from './util'

/**
 * Filter a list of candidate definitions to select those likely to be valid
 * cross-references for a definition in this file. Accept candidates located
 * in the same package or in a directory that includes one of the imported
 * paths.
 *
 * If no candidates match, fall back to the raw (unfiltered) results so that
 * the user doesn't get an empty response unless there really is nothing.
 */
function filterDefinitions<T extends Result>(results: T[], { repo, filePath, fileContent }: FilterContext): T[] {
    const importPaths = extractFromLines(fileContent, /^(?:import |\t)(?:\w+ |\. )?"(.*)"$/)

    return filterResults(results, ({ repo: resultRepo, file }) => {
        const resultImportPath = importPath(resultRepo, file)

        return (
            // Match results from the same package
            resultImportPath === importPath(repo, filePath) ||
            // Match results that are imported explicitly
            importPaths.some(index => resultImportPath.includes(index))
        )
    })
}

/**
 * Return the Go import path for a particular file.
 *
 * @param repo The name of the repository.
 * @param filePath The relative path to the file from the repo root.
 */
function importPath(repo: string, filePath: string): string {
    return `${repo}/${path.dirname(filePath)}`
}

export const goSpec: LanguageSpec = {
    languageID: 'go',
    stylized: 'Go',
    fileExts: ['go'],
    commentStyles: [{ lineRegex: slashPattern }],
    filterDefinitions,
    lsifSupport: LSIFSupport.Robust,
    textDocumentImplemenationSupport: true,
}
