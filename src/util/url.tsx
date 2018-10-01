import { Range } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, PositionSpec, RangeSpec, RenderModeSpec, RepoFile, ViewStateSpec } from '../repo'

function toRenderModeQuery(ctx: Partial<RenderModeSpec>): string {
    if (ctx.renderMode === 'code') {
        return '?view=code'
    }
    return ''
}

/**
 * Represents a line, a position, a line range, or a position range. It forbids
 * just a character, or a range from a line to a position or vice versa (such as
 * "L1-2:3" or "L1:2-3"), none of which would make much sense.
 *
 * 1-indexed.
 */
export type LineOrPositionOrRange =
    | { line?: undefined; character?: undefined; endLine?: undefined; endCharacter?: undefined }
    | { line: number; character?: number; endLine?: undefined; endCharacter?: undefined }
    | { line: number; character?: undefined; endLine?: number; endCharacter?: undefined }
    | { line: number; character: number; endLine: number; endCharacter: number }

/**
 * Tells if the given fragment component is a legacy blob hash component or not.
 *
 * @param hash The URL fragment.
 */
export function isLegacyFragment(hash: string): boolean {
    if (hash.startsWith('#')) {
        hash = hash.substr('#'.length)
    }
    return (
        hash !== '' &&
        !hash.includes('=') &&
        (hash.includes('$info') ||
            hash.includes('$def') ||
            hash.includes('$references') ||
            hash.includes('$references:external') ||
            hash.includes('$impl') ||
            hash.includes('$history'))
    )
}

/**
 * Parses the URL fragment (hash) portion, which consists of a line, position, or range in the file, plus an
 * optional "viewState" parameter (that encodes other view state, such as for the panel).
 *
 * For example, in the URL fragment "#L17:19-21:23$foo:bar", the "viewState" is "foo:bar".
 *
 * @template V The type that describes the view state (typically a union of string constants). There is no runtime
 *             check that the return value satisfies V.
 */
export function parseHash<V extends string>(hash: string): LineOrPositionOrRange & { viewState?: V } {
    if (hash.startsWith('#')) {
        hash = hash.substr('#'.length)
    }

    if (!isLegacyFragment(hash)) {
        // Modern hash parsing logic (e.g. for hashes like `"#L17:19-21:23&tab=foo:bar"`:
        const searchParams = new URLSearchParams(hash)
        const lpr = (findLineInSearchParams(searchParams) || {}) as LineOrPositionOrRange & {
            viewState?: V
        }
        if (searchParams.get('tab')) {
            lpr.viewState = searchParams.get('tab') as V
        }
        return lpr
    }

    // Legacy hash parsing logic (e.g. for hashes like "#L17:19-21:23$foo:bar" where the "viewState" is "foo:bar"):
    if (!/^(L[0-9]+(:[0-9]+)?(-[0-9]+(:[0-9]+)?)?)?(\$.*)?$/.test(hash)) {
        // invalid or empty hash
        return {}
    }
    const lineCharModalInfo = hash.split('$', 2) // e.g. "L17:19-21:23$references:external"
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
    if (!/^(L[0-9]+(:[0-9]+)?(-[0-9]+(:[0-9]+)?)?)?$/.test(lineChar)) {
        return {} // invalid
    }

    // Parse the line or position range, ensuring we don't get an inconsistent result
    // (such as L1-2:3, a range from a line to a position).
    let line: number | undefined // 17
    let character: number | undefined // 19
    let endLine: number | undefined // 21
    let endCharacter: number | undefined // 23
    if (lineChar.startsWith('L')) {
        const posOrRangeString = lineChar.slice(1)
        const [startString, endString] = posOrRangeString.split('-', 2)
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
 * Returns the LineOrPositionOrRange and given URLSearchParams as a string.
 */
export function formatHash(lpr: LineOrPositionOrRange, searchParams: URLSearchParams): string {
    if (!lpr.line) {
        return `#${searchParams.toString()}`
    }
    const anyParams = Array.from(searchParams).length > 0
    return `#L${formatLineOrPositionOrRange(lpr)}${anyParams ? '&' + searchParams.toString() : ''}`
}

/**
 * Returns the textual form of the LineOrPositionOrRange suitable for encoding
 * in a URL fragment' query parameter.
 *
 * @param lpr The `LineOrPositionOrRange`
 */
function formatLineOrPositionOrRange(lpr: LineOrPositionOrRange): string {
    const range = lprToRange(lpr)
    if (!range) {
        return ''
    }
    const emptyRange = range.start.line === range.end.line && range.start.character === range.end.character
    return emptyRange
        ? toPositionHashComponent(range.start)
        : `${toPositionHashComponent(range.start)}-${toPositionHashComponent(range.end)}`
}

/**
 * Finds the URL search parameter which has a key like "L1-2:3" without any
 * value.
 *
 * @param searchParams The URLSearchParams to look for the line in.
 */
function findLineInSearchParams(searchParams: URLSearchParams): LineOrPositionOrRange | undefined {
    for (const key of searchParams.keys()) {
        if (key.startsWith('L')) {
            return parseLineOrPositionOrRange(key)
        }
        break
    }
    return undefined
}

export function lprToRange(lpr: LineOrPositionOrRange): Range | undefined {
    if (lpr.line === undefined) {
        return undefined
    }
    return {
        start: { line: lpr.line, character: lpr.character || 0 },
        end: {
            line: lpr.endLine || lpr.line,
            character: lpr.endCharacter || lpr.character || 0,
        },
    }
}

function parseLineOrPosition(
    str: string
): { line: undefined; character: undefined } | { line: number; character?: number } {
    const parts = str.split(':', 2)
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
 * @param ctx 1-indexed partial position or range spec
 */
export function toPositionOrRangeHash(ctx: {
    position?: { line: number; character?: number }
    range?: { start: { line: number; character?: number }; end: { line: number; character?: number } }
}): string {
    if (ctx.range) {
        const emptyRange =
            ctx.range.start.line === ctx.range.end.line && ctx.range.start.character === ctx.range.end.character
        return (
            '#L' +
            (emptyRange
                ? toPositionHashComponent(ctx.range.start)
                : `${toPositionHashComponent(ctx.range.start)}-${toPositionHashComponent(ctx.range.end)}`)
        )
    }
    if (ctx.position) {
        return '#L' + toPositionHashComponent(ctx.position)
    }
    return ''
}

/**
 * @param ctx 1-indexed partial position
 */
function toPositionHashComponent(position: { line: number; character?: number }): string {
    return position.line.toString() + (position.character ? ':' + position.character : '')
}

/** Encodes a repository at a revspec for use in a URL. */
export function encodeRepoRev(repo: string, rev?: string): string {
    return rev ? `${repo}@${escapeRevspecForURL(rev)}` : repo
}

/**
 * Encodes rev with encodeURIComponent, except that slashes ('/') are preserved,
 * because they are not ambiguous in any of the current places where used, and URLs
 * for (e.g.) branches with slashes look a lot nicer with '/' than '%2F'.
 */
export function escapeRevspecForURL(rev: string): string {
    return encodeURIComponent(rev).replace(/%2F/g, '/')
}

export function toViewStateHashComponent(viewState: string | undefined): string {
    return viewState ? `&tab=${viewState}` : ''
}

/**
 * Returns the URL path for the given repository name.
 *
 * @deprecated Obtain the repository's URL from the GraphQL Repository.url field instead.
 */
export function toRepoURL(repoName: string): string {
    return `/${repoName}`
}

export function toPrettyBlobURL(
    ctx: RepoFile & Partial<PositionSpec> & Partial<ViewStateSpec> & Partial<RangeSpec> & Partial<RenderModeSpec>
): string {
    return `/${encodeRepoRev(ctx.repoPath, ctx.rev)}/-/blob/${ctx.filePath}${toRenderModeQuery(
        ctx
    )}${toPositionOrRangeHash(ctx)}${toViewStateHashComponent(ctx.viewState)}`
}

export function toAbsoluteBlobURL(
    ctx: AbsoluteRepoFile & Partial<PositionSpec> & Partial<ViewStateSpec> & Partial<RenderModeSpec>
): string {
    const rev = ctx.commitID ? ctx.commitID : ctx.rev
    return `/${encodeRepoRev(ctx.repoPath, rev)}/-/blob/${ctx.filePath}${toPositionOrRangeHash(
        ctx
    )}${toViewStateHashComponent(ctx.viewState)}`
}

export function toTreeURL(ctx: RepoFile): string {
    const rev = ctx.commitID || ctx.rev || ''
    return `/${encodeRepoRev(ctx.repoPath, rev)}/-/tree/${ctx.filePath}`
}

/**
 * Correctly handle use of meta/ctrl/alt keys during onClick events that open new pages
 */
export function openFromJS(path: string, event?: MouseEvent): void {
    if (event && (event.metaKey || event.altKey || event.ctrlKey || event.which === 2)) {
        window.open(path, '_blank')
    } else {
        window.location.href = path
    }
}
