import { findLineKeyInSearchParameters, LineOrPositionOrRange, ParsedRepoURI, parseRepoURI } from '@sourcegraph/common'

import { WorkspaceRootWithMetadata } from '../api/extension/extensionHostApi'

/**
 * Tells if the given fragment component is a legacy blob hash component or not.
 *
 * @param hash The URL fragment.
 */
export function isLegacyFragment(hash: string): boolean {
    if (hash.startsWith('#')) {
        hash = hash.slice('#'.length)
    }
    return (
        hash !== '' &&
        !hash.includes('=') &&
        (hash.includes('$info') ||
            hash.includes('$def') ||
            hash.includes('$references') ||
            hash.includes('$impl') ||
            hash.includes('$history'))
    )
}

/**
 * Parses the URL search (query) portion and looks for a parameter which matches a line, position, or range in the file. If not found, it
 * falls back to parsing the hash for backwards compatibility.
 *
 * @template V The type that describes the view state (typically a union of string constants). There is no runtime check that the return value satisfies V.
 */
export function parseQueryAndHash<V extends string>(
    query: string,
    hash: string
): LineOrPositionOrRange & { viewState?: V } {
    const lpr = findLineInSearchParameters(new URLSearchParams(query))
    const parsedHash = parseHash<V>(hash)
    if (!lpr) {
        return parsedHash
    }
    return { ...lpr, viewState: parsedHash.viewState }
}

/**
 * Parses the URL fragment (hash) portion, which consists of a line, position, or range in the file, plus an
 * optional "viewState" parameter (that encodes other view state, such as for the panel).
 *
 * For example, in the URL fragment "#L17:19-21:23$foo:bar", the "viewState" is "foo:bar".
 *
 * @template V The type that describes the view state (typically a union of string constants). There is no runtime check that the return value satisfies V.
 */
export function parseHash<V extends string>(hash: string): LineOrPositionOrRange & { viewState?: V } {
    if (hash.startsWith('#')) {
        hash = hash.slice('#'.length)
    }

    if (!isLegacyFragment(hash)) {
        // Modern hash parsing logic (e.g. for hashes like `"#L17:19-21:23&tab=foo:bar"`:
        const searchParameters = new URLSearchParams(hash)
        const lpr = (findLineInSearchParameters(searchParameters) || {}) as LineOrPositionOrRange & {
            viewState?: V
        }
        if (searchParameters.get('tab')) {
            lpr.viewState = searchParameters.get('tab') as V
        }
        return lpr
    }

    // Legacy hash parsing logic (e.g. for hashes like "#L17:19-21:23$foo:bar" where the "viewState" is "foo:bar"):
    if (!/^(L\d+(:\d+)?(-\d+(:\d+)?)?)?(\$.*)?$/.test(hash)) {
        // invalid or empty hash
        return {}
    }
    const lineCharModalInfo = hash.split('$', 2) // e.g. "L17:19-21:23$references"
    const lpr = parseLineOrPositionOrRange(lineCharModalInfo[0]) as LineOrPositionOrRange & { viewState?: V }
    if (lineCharModalInfo[1]) {
        lpr.viewState = lineCharModalInfo[1] as V
    }
    return lpr
}

/**
 * Parses a string like "L1-2:3", a range from a line to a position.
 */
function parseLineOrPositionOrRange(lineChar: string): LineOrPositionOrRange {
    if (!/^(L\d+(:\d+)?(-L?\d+(:\d+)?)?)?$/.test(lineChar)) {
        return {} // invalid
    }

    // Parse the line or position range, ensuring we don't get an inconsistent result
    // (such as L1-2:3, a range from a line to a position).
    let line: number | undefined // 17
    let character: number | undefined // 19
    let endLine: number | undefined // 21
    let endCharacter: number | undefined // 23
    if (lineChar.startsWith('L')) {
        const positionOrRangeString = lineChar.slice(1)
        const [startString, endString] = positionOrRangeString.split('-', 2)
        if (startString) {
            const parsed = parseLineOrPosition(startString)
            line = parsed.line
            character = parsed.character
        }
        if (endString) {
            const parsed = parseLineOrPosition(endString)
            endLine = parsed.line
            endCharacter = parsed.character
        }
    }
    let lpr = { line, character, endLine, endCharacter } as LineOrPositionOrRange
    if (typeof line === 'undefined' || (typeof endLine !== 'undefined' && typeof character !== typeof endCharacter)) {
        lpr = {}
    } else if (typeof character === 'undefined') {
        lpr = typeof endLine === 'undefined' ? { line } : { line, endLine }
    } else if (typeof endLine === 'undefined' || typeof endCharacter === 'undefined') {
        lpr = { line, character }
    } else {
        lpr = { line, character, endLine, endCharacter }
    }
    return lpr
}

/**
 * Finds the URL search parameter which has a key like "L1-2:3" without any
 * value.
 *
 * @param searchParameters The URLSearchParams to look for the line in.
 */
function findLineInSearchParameters(searchParameters: URLSearchParams): LineOrPositionOrRange | undefined {
    const key = findLineKeyInSearchParameters(searchParameters)
    return key ? parseLineOrPositionOrRange(key) : undefined
}

function parseLineOrPosition(
    string: string
): { line: undefined; character: undefined } | { line: number; character?: number } {
    if (string.startsWith('L')) {
        string = string.slice(1)
    }
    const parts = string.split(':', 2)
    let line: number | undefined
    let character: number | undefined
    if (parts.length >= 1) {
        line = parseInt(parts[0], 10)
    }
    if (parts.length === 2) {
        character = parseInt(parts[1], 10)
    }
    line = typeof line === 'number' && isNaN(line) ? undefined : line
    character = typeof character === 'number' && isNaN(character) ? undefined : character
    if (typeof line === 'undefined') {
        return { line: undefined, character: undefined }
    }
    return { line, character }
}

/**
 * Translate a URI to use the input revision (e.g., branch names) instead of the Git commit SHA if the URI is
 * inside of a workspace root. This helper is used to translate URLs (from actions such as go-to-definition) to
 * avoid navigating the user from (e.g.) a URL with a nice Git branch name to a URL with a full Git commit SHA.
 *
 * For example, suppose there is a workspace root `git://r?a9cb9d` with input revision `mybranch`. If {@link uri}
 * is `git://r?a9cb9d#f`, it would be translated to `git://r?mybranch#f`.
 */
export function withWorkspaceRootInputRevision(
    workspaceRoots: readonly WorkspaceRootWithMetadata[],
    uri: ParsedRepoURI
): ParsedRepoURI {
    const inWorkspaceRoot = workspaceRoots.find(root => {
        const rootURI = parseRepoURI(root.uri)
        return rootURI.repoName === uri.repoName && rootURI.revision === uri.revision
    })
    if (inWorkspaceRoot?.inputRevision !== undefined) {
        return { ...uri, commitID: undefined, revision: inWorkspaceRoot.inputRevision }
    }
    return uri // unchanged
}
