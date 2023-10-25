import { isDefined } from '../util/helpers'

import type { Result } from './language-spec'

/**
 * Extract content from each line of the source file. The first match on
 * each line is returned. Each supplied pattern is expected to have exactly
 * one capture group, which is returned on successful match.
 *
 * @param fileContent The content of the source file.
 * @param patterns The regex patterns executed on each source line.
 */
export function extractFromLines(fileContent: string, ...patterns: RegExp[]): string[] {
    const extractMatch = (line: string): string | undefined => {
        for (const pattern of patterns) {
            const match = pattern.exec(line)
            if (match) {
                return match[1]
            }
        }

        return undefined
    }

    return fileContent.split('\n').map(extractMatch).filter(isDefined)
}

/**
 * Filter the given results by calling the given function on each result
 * and importPath pairs. Remove any filters that do not pass the filter
 * for any import path. If the filtered result is empty, this function
 * returns the original input unchanged.
 *
 * @param results A list of results to filter.
 * @param importPaths A list of import paths.
 * @param func The filter function.
 */
export function filterResultsByImports<T extends Result>(
    results: T[],
    importPaths: string[],
    func: (result: T, importPath: string) => boolean
): T[] {
    return filterResults(results, result => importPaths.some(importPath => func(result, importPath)))
}

/**
 * Filter the given results. If the filtered result is empty, this
 * function returns the original input unchanged.
 *
 * @param results A list of results to filter.
 * @param func The filter function.
 */
export function filterResults<T extends Result>(results: T[], func: (result: T) => boolean): T[] {
    const filteredResults = results.filter(result => func(result))

    // If we filtered out all results, fall back to whatever fuzzy
    // results we had in the first place. It's better than nothing.
    return filteredResults.length === 0 ? results : filteredResults
}

export function removeExtension(filePath: string): string {
    return filePath.replace(/\.[^./]+$/, '')
}

export function slashToDot(value: string): string {
    return value.replaceAll('/', '.')
}

export function dotToSlash(value: string): string {
    return value.replaceAll('.', '/')
}
