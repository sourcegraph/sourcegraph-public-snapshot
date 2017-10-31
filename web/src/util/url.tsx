import { Position } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, PositionSpec, RangeSpec, ReferencesModeSpec, Repo, RepoFile, ResolvedRevSpec } from '../repo'

type Modal = 'references'
type ModalMode = 'local' | 'external'

/**
 * Represents a line, a position, a line range, or a position range. It forbids
 * just a character, or a range from a line to a position or vice versa (such as
 * "L1-2:3" or "L1:2-3"), none of which would make much sense.
 */
export type LineOrPositionOrRange =
    { line?: undefined, character?: undefined, endLine?: undefined, endCharacter?: undefined } |
    { line: number, character?: number, endLine?: undefined, endCharacter?: undefined } |
    { line: number, character?: undefined, endLine?: number, endCharacter?: undefined } |
    { line: number, character: number, endLine: number, endCharacter: number }

export function parseHash(hash: string): LineOrPositionOrRange & { modal?: Modal, modalMode?: ModalMode } {
    if (hash.startsWith('#')) {
        hash = hash.substr('#'.length)
    }
    if (!/^L[0-9]+(:[0-9]+)?($|((-[0-9]+(:[0-9]+)?)?($|(\$references($|(:(local|external)$))))))/.test(hash)) {
        // invalid hash
        return {}
    }

    const lineCharModalInfo = hash.split('$') // e.g. "L17:19-21:23$references:external"

    // Parse the line or position range, ensuring we don't get an inconsistent result
    // (such as L1-2:3, a range from a line to a position).
    let line: number | undefined // 17
    let character: number | undefined // 19
    let endLine: number | undefined // 21
    let endCharacter: number | undefined // 23
    if (lineCharModalInfo[0].startsWith('L')) {
        const posOrRangeString = lineCharModalInfo[0].slice(1)
        const [startString, endString] = posOrRangeString.split('-', 2)
        if (startString) {
            ({ line, character } = parseLineOrPosition(startString))
        }
        if (endString) {
            ({ line: endLine, character: endCharacter } = parseLineOrPosition(endString))
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

    if (!lineCharModalInfo[1]) {
        return lpr
    }

    const modalInfo = lineCharModalInfo[1].split(':')
    const modal = modalInfo[0] as Modal // "references"
    const modalMode = modalInfo[1] as ModalMode || 'local' // "external"
    return { ...lpr, modal, modalMode }
}

function parseLineOrPosition(str: string): { line: undefined, character: undefined } | { line: number, character?: number } {
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

function toPositionOrRangeHash(ctx: Partial<PositionSpec> & Partial<RangeSpec>): string {
    if (ctx.range) {
        const emptyRange = ctx.range.start.line === ctx.range.end.line &&
            ctx.range.start.character === ctx.range.end.character
        return '#L' + (emptyRange ? toPositionHashComponent(ctx.range.start) : `${toPositionHashComponent(ctx.range.start)}-${toPositionHashComponent(ctx.range.end)}`)
    }
    if (ctx.position) {
        return '#L' + toPositionHashComponent(ctx.position)
    }
    return ''
}

function toPositionHashComponent(position: Position): string {
    return position.line.toString() + (position.character ? ':' + position.character : '')
}

function toReferencesHash(group: 'local' | 'external' | undefined): string {
    return group ? (group === 'local' ? '$references' : '$references:external') : ''
}

export function toRepoURL(ctx: Repo & Partial<ResolvedRevSpec>): string {
    const rev = ctx.commitID || ctx.rev || ''
    return `/${ctx.repoPath}${rev ? '@' + rev : ''}`
}

export function toPrettyRepoURL(ctx: Repo): string {
    return `/${ctx.repoPath}${ctx.rev ? '@' + ctx.rev : ''}`
}

export function toBlobURL(ctx: RepoFile & Partial<PositionSpec>): string {
    const rev = ctx.commitID || ctx.rev || ''
    return `/${ctx.repoPath}${rev ? '@' + rev : ''}/-/blob/${ctx.filePath}`
}

export function toPrettyBlobURL(ctx: RepoFile & Partial<PositionSpec> & Partial<ReferencesModeSpec> & Partial<RangeSpec>): string {
    return `/${ctx.repoPath}${ctx.rev ? '@' + ctx.rev : ''}/-/blob/${ctx.filePath}${toPositionOrRangeHash(ctx)}${toReferencesHash(ctx.referencesMode)}`
}

export function toAbsoluteBlobURL(ctx: AbsoluteRepoFile & Partial<PositionSpec> & Partial<ReferencesModeSpec>): string {
    const rev = ctx.commitID ? ctx.commitID : ctx.rev
    return `/${ctx.repoPath}${rev ? '@' + rev : ''}/-/blob/${ctx.filePath}${toPositionOrRangeHash(ctx)}${toReferencesHash(ctx.referencesMode)}`
}

export function toTreeURL(ctx: RepoFile): string {
    const rev = ctx.commitID || ctx.rev || ''
    return `/${ctx.repoPath}${rev ? '@' + rev : ''}/-/tree/${ctx.filePath}`
}

export function toEditorURL(repoPath: string, rev?: string, filePath?: string, position?: { line?: number }, thread?: number): string {
    let query = 'repo=' + encodeURIComponent('ssh://git@' + repoPath + '.git')
    query += '&vcs=git'
    if (rev) {
        query += '&revision=' + encodeURIComponent(rev)
    }
    if (filePath) {
        if (filePath.startsWith('/')) {
            filePath = filePath.substr(1)
        }
        query += '&path=' + encodeURIComponent(filePath)
    }
    if (position && position.line) {
        query += '&selection=' + encodeURIComponent('' + position.line)
    }
    if (thread) {
        query += '&thread=' + encodeURIComponent(String(thread))
    }
    return 'https://about.sourcegraph.com/open/#open?' + query
}

/**
 * Correctly handle use of meta/ctrl/alt keys during onClick events that open new pages
 */
export function openFromJS(path: string, event?: MouseEvent): void {
    if (event && (event.metaKey || event.altKey || event.ctrlKey)) {
        window.open(path, '_blank')
    } else {
        window.location.href = path
    }
}
