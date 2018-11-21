import { Position, Range } from '../../../../../shared/src/api/protocol/plainTypes'
import { sourcegraphUrl } from '../util/context'
import { parseHash } from '../util/url'

/**
 * RepoURI is a URI identifing a repository resource, like
 *   - the repository itself: `git://github.com/gorilla/mux`
 *   - the repository at a particular revision: `git://github.com/gorilla/mux?rev`
 *   - a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go
 *   - a line in a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go:3
 *   - a character position in a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go:3,5
 *   - a rangein a file in a repository at an immutable revision: `git://github.com/gorilla/mux?SHA#path/to/file.go:3,5-4,9
 */
type RepoURI = string

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

export interface PositionSpec {
    /**
     * a point in the blob
     */
    position: Position
}

interface RangeSpec {
    /**
     * a range in the blob
     */
    range: Range
}

export interface ReferencesModeSpec {
    /**
     * the mode for the references panel
     */
    referencesMode: 'local' | 'external'
}

/**
 * Properties of a RepoURI (like git://github.com/gorilla/mux#mux.go) or a URL (like https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go)
 */
export interface ParsedRepoURI
    extends RepoSpec,
        Partial<RevSpec>,
        Partial<ResolvedRevSpec>,
        Partial<FileSpec>,
        Partial<PositionSpec>,
        Partial<RangeSpec> {}

/**
 * A repo resolved to an exact commit
 */
export interface AbsoluteRepo extends RepoSpec, Partial<RevSpec>, ResolvedRevSpec {}

/**
 * A file at an exact commit
 */
export interface AbsoluteRepoFile extends RepoSpec, Partial<RevSpec>, ResolvedRevSpec, FileSpec {}

/**
 * A position in file at an exact commit
 */
export interface AbsoluteRepoFilePosition
    extends RepoSpec,
        Partial<RevSpec>,
        ResolvedRevSpec,
        FileSpec,
        PositionSpec,
        Partial<ReferencesModeSpec> {}

interface DiffRepoSpec {
    baseRepoPath: string
    headRepoPath: string
}

interface DiffRevSpec {
    baseRev: string
    headRev: string
}

export interface DiffResolvedRevSpec {
    baseCommitID: string
    headCommitID: string
}

export interface DiffRepoRev extends DiffRepoSpec, DiffRevSpec {}

export interface OpenInSourcegraphProps {
    repoPath: string
    rev: string
    filePath?: string
    commit?: {
        baseRev: string
        headRev: string
    }
    coords?: {
        line: number
        char: number
    }
    fragment?: 'references'
    query?: {
        search?: string
        diff?: {
            rev: string
        }
    }
    withModifierKey?: boolean
}

export interface MaybeDiffSpec {
    isDelta: boolean
    isSplitDiff?: boolean
    isBase?: boolean
}

export interface CodeCell {
    cell: HTMLElement
    eventHandler: HTMLElement
    line: number
    isAddition?: boolean // for diff views
    isDeletion?: boolean // for diff views
}

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

/**
 * Parses the properties of a blob URL.
 */
export function parseBrowserRepoURL(href: string, w: Window = window): ParsedRepoURI {
    const loc = new URL(href, sourcegraphUrl)
    let pathname = loc.pathname.slice(1) // trim leading '/'
    if (pathname.endsWith('/')) {
        pathname = pathname.substr(0, pathname.length - 1) // trim triling '/'
    }

    const indexOfSep = pathname.indexOf('/-/')
    let repoRev: string // e.g. 'github.com/gorilla/mux' or 'github.com/gorilla/mux@revision'
    if (indexOfSep === -1) {
        repoRev = pathname // the whole string
    } else {
        repoRev = pathname.substring(0, indexOfSep) // the whole string leading up to the separator (allows rev to be multiple path parts)
    }
    const repoRevSplit = repoRev.split('@')
    const repoPath = repoRevSplit[0]
    if (!repoPath) {
        throw new Error('unexpected repo url: ' + href)
    }
    const rev: string | undefined = repoRevSplit[1]
    const commitID = rev && rev.match(/^[a-f0-9]{40}$/i) ? rev : undefined

    let filePath: string | undefined
    const treeSep = pathname.indexOf('/-/tree/')
    const blobSep = pathname.indexOf('/-/blob/')
    if (treeSep !== -1) {
        filePath = pathname.substr(treeSep + '/-/tree/'.length)
    }
    if (blobSep !== -1) {
        filePath = pathname.substr(blobSep + '/-/blob/'.length)
    }

    let position: Position | undefined
    if (loc.hash) {
        const parsedHash = parseHash(loc.hash.substr('#'.length))
        if (parsedHash.line) {
            position = {
                line: parsedHash.line,
                character: parsedHash.character || 0,
            }
        }
    }

    return { repoPath, rev, commitID, filePath, position }
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
