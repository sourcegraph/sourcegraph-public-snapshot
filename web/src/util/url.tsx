import { AbsoluteRepoFilePosition, Position, RepoFile, RepoFilePosition } from 'sourcegraph/repo'

type Modal = 'references'
type ModalMode = 'local' | 'external'

export function parseHash(hash: string): { line?: number, char?: number, modal?: Modal, modalMode?: ModalMode } {
    let line: number | undefined
    let char: number | undefined
    let modal: Modal | undefined
    let modalMode: ModalMode | undefined

    const lineCharModalInfo = hash.split('$') // e.g. "L17:19$references:external"
    if (lineCharModalInfo[0]) {
        const lineChar = lineCharModalInfo[0].split('L')
        if (lineChar[1]) {
            const coords = lineChar[1].split(':')
            line = parseInt(coords[0], 10) // 17
            char = parseInt(coords[1], 10) // 19
        }
    }
    if (lineCharModalInfo[1]) {
        const modalInfo = lineCharModalInfo[1].split(':')
        // TODO(john): validation
        modal = modalInfo[0] as Modal // "references"
        modalMode = modalInfo[1] as ModalMode || 'local' // "external"
    }
    return { line, char, modal, modalMode }
}

export function toPositionHash(position: Position): string {
    return '#L' + position.line + (position.char ? ':' + position.char : '')
}

export function toReferencesHash(group: 'local' | 'external' | undefined): string {
    return group ? (group === 'local' ? '$references' : '$references:external') : ''
}

export function toBlobURL(ctx: RepoFile): string {
    const rev = ctx.commitID || ctx.rev || ''
    return `/${ctx.repoPath}${rev ? '@' + rev : ''}/-/blob/${ctx.filePath}`
}

export function toPrettyBlobPositionURL(ctx: RepoFilePosition): string {
    return `/${ctx.repoPath}${ctx.rev ? '@' + ctx.rev : ''}/-/blob/${ctx.filePath}${toPositionHash(ctx.position)}${toReferencesHash(ctx.referencesMode)}`
}

export function toBlobPositionURL(ctx: AbsoluteRepoFilePosition): string {
    const rev = ctx.commitID ? ctx.commitID : ctx.rev
    return `/${ctx.repoPath}${rev ? '@' + rev : ''}/-/blob/${ctx.filePath}${toPositionHash(ctx.position)}${toReferencesHash(ctx.referencesMode)}`
}

export function toTreeURL(ctx: RepoFile): string {
    const rev = ctx.commitID || ctx.rev || ''
    return `/${ctx.repoPath}${rev ? '@' + rev : ''}/-/tree/${ctx.filePath}`
}

export function toEditorURL(repoPath: string, rev?: string, filePath?: string): string {
    let query = 'repo=' + encodeURIComponent('git+ssh://' + repoPath + '.git')
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
    return 'https://about.sourcegraph.com/open-native/#open?' + query
}

/**
 * Correctly handle use of meta/ctrl/alt keys during onClick events that open new pages
 */
export function openFromJS(path: string, event?: React.MouseEvent<HTMLElement>): void {
    if (event && (event.metaKey || event.altKey || event.ctrlKey)) {
        window.open(path, '_blank')
    } else {
        window.location.href = path
    }
}
