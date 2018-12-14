import { Position } from '@sourcegraph/extension-api-types'
import { AbsoluteRepoFile, PositionSpec } from '../../../../../shared/src/util/url'
import { repoUrlCache, sourcegraphUrl } from './context'

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
