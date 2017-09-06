import * as URI from 'urijs'

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

export interface Position {
    line: number
    char?: number
}

export interface Range {
    start: Position
    end: Position
}

export interface ParsedRepoURI {
    repoPath: string
    rev?: string
    commitID?: string
    filePath?: string
    position?: Position
    range?: Range
}

const parsePosition = (str: string): Position => {
    const split = str.split('.')
    if (split.length === 1) {
        return { line: parseInt(str, 10) }
    }
    if (split.length === 2) {
        return { line: parseInt(split[0], 10), char: parseInt(split[1], 10) }
    }
    throw new Error('unexpected position: ' + str)
}

export function parseRepoURI(uri: RepoURI): ParsedRepoURI {
    const parsed = URI.parse(uri)
    const repoPath = parsed.hostname + '/' + parsed.path
    const rev = parsed.query || undefined
    let commitID: string | undefined
    if (rev && rev.match(/[0-9a-fA-f]{40}/)) {
        commitID = rev
    }
    const fragmentSplit = parsed.fragment.split(':')
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
        throw new Error('unexpected fragment: ' + parsed.fragment)
    }

    return { repoPath, rev, commitID, filePath, position, range }
}

const positionStr = (pos: Position) => pos.line + '' + (pos.char ? ',' + pos.char : '')

export function makeRepoURI(parsed: ParsedRepoURI): RepoURI {
    const rev = parsed.commitID || parsed.rev
    let uri = `git://${parsed.repoPath}`
    uri += rev ? '?' + rev : ''
    uri += parsed.filePath ? '#' + parsed.filePath : ''
    uri += parsed.position ? positionStr(parsed.position) : ''
    uri += parsed.range ? positionStr(parsed.range.start) + '-' + positionStr(parsed.range.end) : ''
    return uri
}
