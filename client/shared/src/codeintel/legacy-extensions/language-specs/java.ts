import * as path from 'path'

import { javaStyleComment } from './comments'
import { FilterContext, LanguageSpec, Result, LSIFSupport } from './language-spec'
import { extractFromLines, filterResultsByImports, slashToDot } from './util'

/**
 * Filter a list of candidate definitions to select those likely to be valid
 * cross-references for a definition in this file. Accept candidates located
 * in the same package or in a directory that includes one of the imported
 * paths.
 *
 * If no candidates match, fall back to the raw (unfiltered) results so that
 * the user doesn't get an empty response unless there really is nothing.
 */
function filterDefinitions<T extends Result>(results: T[], { fileContent }: FilterContext): T[] {
    const importPaths = extractFromLines(
        fileContent,
        /^import ([\w.]+);$/,
        /^import static ([\d._a-z]+)\.([\d_a-z]+|\*);$/
    )

    const currentPackage = extractFromLines(fileContent, /^package ([\w.]+);$/)

    return filterResultsByImports(results, importPaths.concat(currentPackage), ({ file }, importPath) =>
        // Match results with a dirname suffix of an import path
        slashToDot(path.dirname(file)).endsWith(importPath)
    )
}

export const javaSpec: LanguageSpec = {
    languageID: 'java',
    stylized: 'Java',
    fileExts: ['java'],
    commentStyles: [javaStyleComment],
    textDocumentImplemenationSupport: true,
    filterDefinitions,
    lsifSupport: LSIFSupport.Experimental,
}
