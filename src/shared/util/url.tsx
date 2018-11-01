import { Position } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, PositionSpec, ReferencesModeSpec, Repo, RepoFile, ResolvedRevSpec } from '../repo'
import { repoUrlCache, sourcegraphUrl } from './context'

type Modal = 'references'
type ModalMode = 'local' | 'external'

export function parseHash(hash: string): { line?: number; character?: number; modal?: Modal; modalMode?: ModalMode } {
    if (hash.startsWith('#')) {
        hash = hash.substr('#'.length)
    }
    if (hash.startsWith('_')) {
        hash = hash.substr('_'.length)
    }
    if (!/^L[0-9]+($|(:[0-9]+($|(\$references($|(:(local|external)$))))))/.test(hash)) {
        // invalid hash
        return {}
    }

    const lineCharModalInfo = hash.split('$') // e.g. "L17:19$references:external"
    const lineChar = lineCharModalInfo[0].split('L')
    const coords = lineChar[1].split(':')
    const line = parseInt(coords[0], 10) // 17
    const character = coords[1] ? parseInt(coords[1], 10) : undefined // 19

    if (!lineCharModalInfo[1]) {
        return { line, character }
    }

    const modalInfo = lineCharModalInfo[1].split(':')
    const modal = modalInfo[0] as Modal // "references"
    const modalMode = (modalInfo[1] as ModalMode) || 'local' // "external"
    return { line, character, modal, modalMode }
}

function toPositionHash(position?: Position): string {
    if (!position) {
        return ''
    }
    return '#L' + position.line + (position.character ? ':' + position.character : '')
}

function toReferencesHash(group: 'local' | 'external' | undefined): string {
    return group ? (group === 'local' ? '$references' : '$references:external') : ''
}

export function toRepoURL(ctx: Repo & Partial<ResolvedRevSpec>): string {
    const url = repoUrlCache[ctx.repoPath] || sourcegraphUrl
    const rev = ctx.commitID || ctx.rev || ''
    return `${url}/${ctx.repoPath}${rev ? '@' + rev : ''}`
}

export function toPrettyRepoURL(ctx: Repo): string {
    const url = repoUrlCache[ctx.repoPath] || sourcegraphUrl
    return `${url}/${ctx.repoPath}${ctx.rev ? '@' + ctx.rev : ''}`
}

export function toBlobURL(ctx: RepoFile & Partial<PositionSpec>): string {
    const url = repoUrlCache[ctx.repoPath] || sourcegraphUrl
    const rev = ctx.commitID || ctx.rev || ''
    return `${url}/${ctx.repoPath}${rev ? '@' + rev : ''}/-/blob/${ctx.filePath}`
}

export function toPrettyBlobURL(ctx: RepoFile & Partial<PositionSpec> & Partial<ReferencesModeSpec>): string {
    const url = repoUrlCache[ctx.repoPath] || sourcegraphUrl
    return `${url}/${ctx.repoPath}${ctx.rev ? '@' + ctx.rev : ''}/-/blob/${ctx.filePath}${toPositionHash(
        ctx.position
    )}${toReferencesHash(ctx.referencesMode)}`
}

export function toAbsoluteBlobURL(ctx: AbsoluteRepoFile & Partial<PositionSpec> & Partial<ReferencesModeSpec>): string {
    const rev = ctx.commitID ? ctx.commitID : ctx.rev
    const url = repoUrlCache[ctx.repoPath] || sourcegraphUrl

    return `${url}/${ctx.repoPath}${rev ? '@' + rev : ''}/-/blob/${ctx.filePath}${toPositionHash(
        ctx.position
    )}${toReferencesHash(ctx.referencesMode)}`
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

/**
 * Builds a URL query for given SearchOptions (without leading `?`)
 */
export function buildSearchURLQuery(query: string): string {
    const searchParams = new URLSearchParams()
    searchParams.set('q', query)
    return searchParams
        .toString()
        .replace(/%2F/g, '/')
        .replace(/%3A/g, ':')
}
