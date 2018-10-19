import { Position, Range } from 'vscode-languageserver-types'
import * as url from '../util/url'

/**
 * RepoURI is a URI identifing a repository resource, like
 *   - the repository itself: `git://github.com/gorilla/mux`
 *   - the repository at a particular revision: `git://github.com/gorilla/mux?rev`
 *   - a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go
 *   - a line in a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go:3
 *   - a character position in a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go:3,5
 *   - a rangein a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go:3,5-4,9
 */
export type RepoURI = string

export interface RepoSpec {
    /**
     * Example: github.com/gorilla/mux
     */
    repoPath: string
}

export interface RevSpec {
    /**
     * a revision string (like 'master' or 'my-branch' or '24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
     */
    rev: string
}

export interface ResolvedRevSpec {
    /**
     * a 40 character commit SHA
     */
    commitID: string
}

export interface FileSpec {
    /**
     * a path to a directory or file
     */
    filePath: string
}

export interface ComparisonSpec {
    /**
     * a diff specifier with optional base and comparison. Examples:
     * - "master..." (implicitly: "master...HEAD")
     * - "...my-branch" (implicitly: "HEAD...my-branch")
     * - "master...my-branch"
     */
    commitRange: string
}

export interface PositionSpec {
    /**
     * a 1-indexed point in the blob
     */
    position: Position
}

export interface RangeSpec {
    /**
     * a 1-indexed range in the blob
     */
    range: Range
}

export type BlobViewState = 'references' | 'references:external' | 'discussions' | 'impl'

export interface ViewStateSpec {
    /**
     * The view state (for the blob panel).
     */
    viewState: BlobViewState
}

/**
 * 'code' for Markdown/rich-HTML files rendered as code, 'rendered' for rendering them as
 * Markdown/rich-HTML, undefined for the default for the file type ('rendered' for Markdown, etc.,
 * 'code' otherwise).
 */
export type RenderMode = 'code' | 'rendered' | undefined

export interface RenderModeSpec {
    /**
     * How the file should be rendered.
     */
    renderMode: RenderMode
}

/**
 * Properties of a RepoURI (like git://github.com/gorilla/mux#mux.go) or a URL (like https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go)
 */
export interface ParsedRepoURI
    extends RepoSpec,
        Partial<RevSpec>,
        Partial<ResolvedRevSpec>,
        Partial<FileSpec>,
        Partial<ComparisonSpec>,
        Partial<PositionSpec>,
        Partial<RangeSpec> {}

/**
 * A repo
 */
export interface Repo extends RepoSpec {}

/**
 * A repo with a (possibly unresolved) revspec.
 */
export interface RepoRev extends RepoSpec, RevSpec {}

/**
 * A repo resolved to an exact commit
 */
export interface AbsoluteRepo extends RepoSpec, RevSpec, ResolvedRevSpec {}

/**
 * A file in a repo
 */
export interface RepoFile extends RepoSpec, RevSpec, Partial<ResolvedRevSpec>, FileSpec {}

/**
 * A file at an exact commit
 */
export interface AbsoluteRepoFile extends RepoSpec, RevSpec, ResolvedRevSpec, FileSpec {}

/**
 * A position in file
 */
export interface RepoFilePosition
    extends RepoSpec,
        RevSpec,
        Partial<ResolvedRevSpec>,
        FileSpec,
        PositionSpec,
        Partial<ViewStateSpec>,
        Partial<RenderModeSpec> {}

/**
 * A position in file at an exact commit
 */
export interface AbsoluteRepoFilePosition
    extends RepoSpec,
        RevSpec,
        ResolvedRevSpec,
        FileSpec,
        PositionSpec,
        Partial<ViewStateSpec>,
        Partial<RenderModeSpec> {}

/**
 * A range in file at an exact commit
 */
export interface AbsoluteRepoFileRange
    extends RepoSpec,
        RevSpec,
        ResolvedRevSpec,
        FileSpec,
        RangeSpec,
        Partial<ViewStateSpec>,
        Partial<RenderModeSpec> {}

const parsePosition = (str: string): Position => {
    const split = str.split(',')
    if (split.length === 1) {
        return { line: parseInt(str, 10), character: 0 }
    }
    if (split.length === 2) {
        return { line: parseInt(split[0], 10), character: parseInt(split[1], 10) }
    }
    throw new Error('unexpected position: ' + str)
}

/**
 * Parses the properties of a repo URI like git://github.com/gorilla/mux#mux.go
 */
export function parseRepoURI(uri: RepoURI): ParsedRepoURI {
    const parsed = new URL(uri)
    const repoPath = parsed.hostname + parsed.pathname
    const rev = parsed.search.substr('?'.length) || undefined
    let commitID: string | undefined
    if (rev && rev.match(/[0-9a-fA-f]{40}/)) {
        commitID = rev
    }
    const fragmentSplit = parsed.hash.substr('#'.length).split(':')
    let filePath: string | undefined
    let position: Position | undefined
    let range: Range | undefined
    if (fragmentSplit.length === 1) {
        filePath = fragmentSplit[0]
    }
    if (fragmentSplit.length === 2) {
        filePath = fragmentSplit[0]
        const rangeOrPosition = fragmentSplit[1]
        const rangeOrPositionSplit = rangeOrPosition.split('-')

        if (rangeOrPositionSplit.length === 1) {
            position = parsePosition(rangeOrPositionSplit[0])
        }
        if (rangeOrPositionSplit.length === 2) {
            range = { start: parsePosition(rangeOrPositionSplit[0]), end: parsePosition(rangeOrPositionSplit[1]) }
        }
        if (rangeOrPositionSplit.length > 2) {
            throw new Error('unexpected range or position: ' + rangeOrPosition)
        }
    }
    if (fragmentSplit.length > 2) {
        throw new Error('unexpected fragment: ' + parsed.hash)
    }

    return { repoPath, rev, commitID, filePath: filePath || undefined, position, range }
}

/** The results of parsing a repo-rev string like "my/repo@my/rev". */
export interface ParsedRepoRev {
    repoPath: string

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
        repoPath: decodeURIComponent(repo),
        rev: rev && decodeURIComponent(rev),
        rawRev: rev,
    }
}

/**
 * Parses the properties of a blob URL.
 */
export function parseBrowserRepoURL(href: string): ParsedRepoURI {
    const loc = new URL(href, window.location.href)
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
    const { repoPath, rev } = parseRepoRev(repoRev)
    if (!repoPath) {
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
        const parsedHash = url.parseHash(loc.hash.substr('#'.length))
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

    return { repoPath, rev, commitID, filePath, commitRange, position, range }
}

/**
 * Replaces the revision in the given URL, or adds one if there is not already
 * one.
 * @param href The URL whose revision should be replaced.
 */
export function replaceRevisionInURL(href: string, newRev: string): string {
    const parsed = parseBrowserRepoURL(window.location.href)
    const repoRev = `/${url.encodeRepoRev(parsed.repoPath, parsed.rev)}`

    const u = new URL(window.location.href)
    u.pathname = `/${url.encodeRepoRev(parsed.repoPath, newRev)}${u.pathname.slice(repoRev.length)}`
    return `${u.pathname}${u.search}${u.hash}`
}

const positionStr = (pos: Position) => pos.line + '' + (pos.character ? ',' + pos.character : '')

/**
 * The inverse of parseRepoURI, this generates a string from parsed values.
 */
export function makeRepoURI(parsed: ParsedRepoURI): RepoURI {
    const rev = parsed.commitID || parsed.rev
    let uri = `git://${parsed.repoPath}`
    uri += rev ? '?' + rev : ''
    uri += parsed.filePath ? '#' + parsed.filePath : ''
    uri += parsed.position || parsed.range ? ':' : ''
    uri += parsed.position ? positionStr(parsed.position) : ''
    uri += parsed.range ? positionStr(parsed.range.start) + '-' + positionStr(parsed.range.end) : ''
    return uri
}

/**
 * Retrieves the <td> element at the specified line on the current document.
 */
export function getBlobTableRow(line: number): HTMLTableRowElement {
    const table = document.querySelector('.blob > table') as HTMLTableElement
    return table.rows[line - 1]
}

/**
 * Retrieves the <td> elements for the specified line range (inclusive) on
 * the current document.
 */
export function getBlobTableRows(line: number, endLine: number = line): HTMLTableRowElement[] {
    const table = document.querySelector('.blob > table') as HTMLTableElement
    const rows: HTMLTableRowElement[] = []
    for (let i = line; i <= endLine; i++) {
        const cell = table.rows[i - 1]
        if (cell) {
            rows.push(cell)
        }
    }
    return rows
}

/**
 * Performs a redirect to the host of the given URL with the path, query etc. properties of the current URL.
 */
export function redirectToExternalHost(externalRedirectURL: string): void {
    const externalHostURL = new URL(externalRedirectURL)
    const redirectURL = new URL(window.location.href)
    // Preserve the path of the current URL and redirect to the repo on the external host.
    redirectURL.host = externalHostURL.host
    redirectURL.port = externalHostURL.port
    redirectURL.protocol = externalHostURL.protocol
    window.location.replace(redirectURL.toString())
}
