import { Position } from '@sourcegraph/extension-api-types'
import { AbsoluteRepoFile, PositionSpec } from '../../../../../shared/src/util/url'
import { repoUrlCache, sourcegraphUrl } from './context'

export function parseHash(hash: string): { line?: number; character?: number } {
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

    const lineCharModalInfo = hash.split('$') // e.g. "L17:19$references"
    const lineChar = lineCharModalInfo[0].split('L')
    const coords = lineChar[1].split(':')
    const line = parseInt(coords[0], 10) // 17
    const character = coords[1] ? parseInt(coords[1], 10) : undefined // 19

    return { line, character }
}

function toPositionHash(position?: Position): string {
    if (!position) {
        return ''
    }
    return '#L' + position.line + (position.character ? ':' + position.character : '')
}

export function toAbsoluteBlobURL(ctx: AbsoluteRepoFile & Partial<PositionSpec>): string {
    const rev = ctx.commitID ? ctx.commitID : ctx.rev
    const url = repoUrlCache[ctx.repoName] || sourcegraphUrl

    return `${url}/${ctx.repoName}${rev ? '@' + rev : ''}/-/blob/${ctx.filePath}${toPositionHash(ctx.position)}`
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
