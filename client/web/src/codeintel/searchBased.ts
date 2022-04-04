import { extname } from 'path'

import escapeRegExp from 'lodash/escapeRegExp'

import { SettingsGetter } from './settings'

export function definitionQuery({
    searchToken,
    path,
    fileExts,
}: {
    /** The search token text. */
    searchToken: string
    /** The path to file **/
    path: string
    /** File extensions used by the current extension. */
    fileExts: string[]
}): string[] {
    return [
        `^${searchToken}$`,
        'type:symbol',
        'patternType:regexp',
        'count:500',
        'case:yes',
        fileExtensionTerm(path, fileExts),
    ]
}

/**
 * Create a search query to find references of a symbol.
 *
 * @param args Parameter bag.
 */
export function referencesQuery({
    searchToken,
    path,
    fileExts,
}: {
    /** The search token text. */
    searchToken: string
    /** The path to file **/
    path: string
    /** File extensions used by the current extension. */
    fileExts: string[]
}): string[] {
    let pattern = ''
    if (/^\w/.test(searchToken)) {
        pattern += '\\b'
    }
    pattern += escapeRegExp(searchToken)
    if (/\w$/.test(searchToken)) {
        pattern += '\\b'
    }
    return [pattern, 'type:file', 'patternType:regexp', 'count:500', 'case:yes', fileExtensionTerm(path, fileExts)]
}
/**
 * Constructs a file term containing include-listed extensions. If the current
 * text document path has an excluded extension or an extension absent from the
 * include list, an empty file term will be returned.
 *
 * @param textDocument The current text document.
 * @param includelist The file extensions for the current language.
 */
function fileExtensionTerm(path: string, includelist: string[]): string {
    const extension = extname(path).slice(1)
    if (!extension || excludelist.has(extension) || !includelist.includes(extension)) {
        return ''
    }

    return `file:\\.(${includelist.join('|')})$`
}

const excludelist = new Set(['thrift', 'proto', 'graphql'])

/**
 * Returns fork and archived terms that should be supplied with the query.
 *
 * @param includeFork Whether or not the include forked repositories regardless of settings.
 * @param includeArchived Whether or not the include archived repositories regardless of settings.
 * @param getSetting Used to query user settings for code intel configuration.
 */
export function repositoryKindTerms(
    includeFork: boolean,
    includeArchived: boolean,
    getSetting: SettingsGetter
): string[] {
    const additionalTerms = []
    if (includeFork || getSetting('basicCodeIntel.includeForks', false)) {
        additionalTerms.push('fork:yes')
    }

    if (includeArchived || getSetting('basicCodeIntel.includeArchives', false)) {
        additionalTerms.push('archived:yes')
    }

    return additionalTerms
}
