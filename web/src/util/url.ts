import { Position, Range } from '@sourcegraph/extension-api-types'
import {
    AbsoluteRepoFile,
    encodeRepoRev,
    LineOrPositionOrRange,
    lprToRange,
    ParsedRepoURI,
    parseHash,
    PositionSpec,
    RenderModeSpec,
    RepoFile,
    toPositionHashComponent,
    toPositionOrRangeHash,
    toViewStateHashComponent,
    ViewStateSpec,
} from '../../../shared/src/util/url'

export function toTreeURL(ctx: RepoFile): string {
    const rev = ctx.commitID || ctx.rev || ''
    return `/${encodeRepoRev(ctx.repoName, rev)}/-/tree/${ctx.filePath}`
}

export function toAbsoluteBlobURL(
    ctx: AbsoluteRepoFile & Partial<PositionSpec> & Partial<ViewStateSpec> & Partial<RenderModeSpec>
): string {
    const rev = ctx.commitID ? ctx.commitID : ctx.rev
    return `/${encodeRepoRev(ctx.repoName, rev)}/-/blob/${ctx.filePath}${toPositionOrRangeHash(
        ctx
    )}${toViewStateHashComponent(ctx.viewState)}`
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
 * Replaces the revision in the given URL, or adds one if there is not already
 * one.
 * @param href The URL whose revision should be replaced.
 */
export function replaceRevisionInURL(href: string, newRev: string): string {
    const parsed = parseBrowserRepoURL(window.location.href)
    const repoRev = `/${encodeRepoRev(parsed.repoName, parsed.rev)}`

    const u = new URL(window.location.href)
    u.pathname = `/${encodeRepoRev(parsed.repoName, newRev)}${u.pathname.slice(repoRev.length)}`
    return `${u.pathname}${u.search}${u.hash}`
}

/**
 * Parses the properties of a blob URL.
 */
export function parseBrowserRepoURL(href: string): ParsedRepoURI {
    const loc = new URL(href, typeof window !== 'undefined' ? window.location.href : undefined)
    let pathname = loc.pathname.slice(1) // trim leading '/'
    if (pathname.endsWith('/')) {
        pathname = pathname.substr(0, pathname.length - 1) // trim trailing '/'
    }

    const indexOfSep = pathname.indexOf('/-/')

    // examples:
    // - 'github.com/gorilla/mux'
    // - 'github.com/gorilla/mux@revision'
    // - 'foo/bar' (from 'sourcegraph.mycompany.com/foo/bar')
    // - 'foo/bar@revision' (from 'sourcegraph.mycompany.com/foo/bar@revision')
    // - 'foobar' (from 'sourcegraph.mycompany.com/foobar')
    // - 'foobar@revision' (from 'sourcegraph.mycompany.com/foobar@revision')
    let repoRev: string
    if (indexOfSep === -1) {
        repoRev = pathname // the whole string
    } else {
        repoRev = pathname.substring(0, indexOfSep) // the whole string leading up to the separator (allows rev to be multiple path parts)
    }
    const { repoName, rev } = parseRepoRev(repoRev)
    if (!repoName) {
        throw new Error('unexpected repo url: ' + href)
    }
    const commitID = rev && /^[a-f0-9]{40}$/i.test(rev) ? rev : undefined

    let filePath: string | undefined
    let commitRange: string | undefined
    const treeSep = pathname.indexOf('/-/tree/')
    const blobSep = pathname.indexOf('/-/blob/')
    const comparisonSep = pathname.indexOf('/-/compare/')
    if (treeSep !== -1) {
        filePath = pathname.substr(treeSep + '/-/tree/'.length)
    }
    if (blobSep !== -1) {
        filePath = pathname.substr(blobSep + '/-/blob/'.length)
    }
    if (comparisonSep !== -1) {
        commitRange = pathname.substr(comparisonSep + '/-/compare/'.length)
    }
    let position: Position | undefined
    let range: Range | undefined
    if (loc.hash) {
        const parsedHash = parseHash(loc.hash.substr('#'.length))
        if (parsedHash.line) {
            position = {
                line: parsedHash.line,
                character: parsedHash.character || 0,
            }
            if (parsedHash.endLine) {
                range = {
                    start: position,
                    end: {
                        line: parsedHash.endLine,
                        character: parsedHash.endCharacter || 0,
                    },
                }
            }
        }
    }

    return { repoName, rev, commitID, filePath, commitRange, position, range }
}

/** The results of parsing a repo-rev string like "my/repo@my/rev". */
export interface ParsedRepoRev {
    repoName: string

    /** The URI-decoded revision (e.g., "my#branch" in "my/repo@my%23branch"). */
    rev?: string

    /** The raw revision (e.g., "my%23branch" in "my/repo@my%23branch"). */
    rawRev?: string
}

/**
 * Parses a repo-rev string like "my/repo@my/rev" to the repo and rev components.
 */
export function parseRepoRev(repoRev: string): ParsedRepoRev {
    const [repo, rev] = repoRev.split('@', 2)
    return {
        repoName: decodeURIComponent(repo),
        rev: rev && decodeURIComponent(rev),
        rawRev: rev,
    }
}

/**
 * Correctly handle use of meta/ctrl/alt keys during onClick events that open new pages
 */
export function openFromJS(path: string, event?: MouseEvent): void {
    if (event && (event.metaKey || event.altKey || event.ctrlKey || event.button === 1)) {
        window.open(path, '_blank')
    } else {
        window.location.href = path
    }
}
