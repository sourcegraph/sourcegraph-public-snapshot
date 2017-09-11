import { memoizedFetch } from 'sourcegraph/backend'
import { doFetch as fetch } from 'sourcegraph/backend/xhr'
import { Reference } from 'sourcegraph/references'
import { AbsoluteRepo, AbsoluteRepoFile, AbsoluteRepoFilePosition, makeRepoURI } from 'sourcegraph/repo'
import { getModeFromExtension, getPathExtension, supportedExtensions } from 'sourcegraph/util'

interface LSPRequest {
    method: string
    params: any
}

function wrapLSP(req: LSPRequest, ctx: AbsoluteRepo, path: string): any[] {
    return [
        {
            id: 0,
            method: 'initialize',
            params: {
                // TODO(sqs): rootPath is deprecated but xlang client proxy currently
                // requires it. Pass rootUri as well (below) for forward compat.
                rootPath: `git://${ctx.repoPath}?${ctx.commitID}`,

                rootUri: `git://${ctx.repoPath}?${ctx.commitID}`,
                mode: `${getModeFromExtension(getPathExtension(path))}`
            }
        },
        {
            id: 1,
            ...req
        },
        {
            id: 2,
            method: 'shutdown'
        },
        {
            // id not included on 'exit' requests
            method: 'exit'
        }
    ]
}

export interface Tooltip {
    title?: string
    doc?: string
}

export const EEMPTYTOOLTIP = 'EEMPTYTOOLTIP'

export const getTooltip = memoizedFetch((pos: AbsoluteRepoFilePosition): Promise<Tooltip> => {
    const ext = getPathExtension(pos.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve({})
    }

    const body = wrapLSP({
        method: 'textDocument/hover',
        params: {
            textDocument: {
                uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`
            },
            position: {
                character: pos.position.char! - 1,
                line: pos.position.line - 1
            }
        }
    }, pos, pos.filePath)

    return fetch(`/.api/xlang/textDocument/hover`, { method: 'POST', body: JSON.stringify(body) })
        .then(resp => resp.json())
        .then(json => {
            if (!json[1] ||
                !json[1].result ||
                !json[1].result.contents ||
                json[1].result.contents.length === 0) {
                throw Object.assign(new Error('empty tooltip'), { code: EEMPTYTOOLTIP })
            }
            const title: string = json[1].result.contents[0].value
            let doc: string | undefined
            for (const markedString of json[1].result.contents) {
                if (typeof markedString === 'string') {
                    doc = markedString
                } else if (markedString.language === 'markdown') {
                    doc = markedString.value
                }
            }
            return { title, doc }
        })
}, makeRepoURI)

export const fetchJumpURL = memoizedFetch((pos: AbsoluteRepoFilePosition): Promise<string | null> => {
    const ext = getPathExtension(pos.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve(null)
    }

    const body = wrapLSP({
        method: 'textDocument/definition',
        params: {
            textDocument: {
                uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`
            },
            position: {
                character: pos.position.char! - 1,
                line: pos.position.line - 1
            }
        }
    }, pos, pos.filePath)

    return fetch(`/.api/xlang/textDocument/definition`, { method: 'POST', body: JSON.stringify(body) })
        .then(resp => resp.json())
        .then(json => {
            if (!json ||
                !json[1] ||
                !json[1].result ||
                !json[1].result[0] ||
                !json[1].result[0].uri) {
                // TODO(john): better error handling.
                return null
            }
            const respUri = json[1].result[0].uri.split('git://')[1]
            const prt0Uri = respUri.split('?')
            const prt1Uri = prt0Uri[1].split('#')

            const repoUri = prt0Uri[0]
            let frevUri = repoUri === pos.repoPath ? pos.rev || pos.commitID : prt1Uri[0]
            if (frevUri) {
                frevUri = `@${frevUri}`
            }
            const pathUri = prt1Uri[1]
            const startLine = parseInt(json[1].result[0].range.start.line, 10) + 1
            const startChar = parseInt(json[1].result[0].range.start.character, 10) + 1

            let lineAndCharEnding = ''
            if (startLine && startChar) {
                lineAndCharEnding = `#L${startLine}:${startChar}`
            } else if (startLine) {
                lineAndCharEnding = `#L${startLine}`
            }

            return `/${repoUri}${frevUri || ''}/-/blob/${pathUri}${lineAndCharEnding}`
        })
}, makeRepoURI)

export const fetchXdefinition = memoizedFetch((pos: AbsoluteRepoFilePosition): Promise<{ location: any, symbol: any } | null> => {
    const body = wrapLSP({
        method: 'textDocument/xdefinition',
        params: {
            textDocument: {
                uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`
            },
            position: {
                character: pos.position.char! - 1,
                line: pos.position.line - 1
            }
        }
    }, pos, pos.filePath)

    return fetch(`/.api/xlang/textDocument/xdefinition`, { method: 'POST', body: JSON.stringify(body) })
        .then(resp => resp.json())
        .then(json => {
            if (!json ||
                !json[1] ||
                !json[1].result ||
                !json[1].result[0]) {
                return null
            }

            return json[1].result[0]
        })
}, makeRepoURI)

export const fetchReferences = memoizedFetch((ctx: AbsoluteRepoFilePosition): Promise<Reference[]> => {
    const ext = getPathExtension(ctx.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve([])
    }
    const body = wrapLSP({
        method: 'textDocument/references',
        params: {
            textDocument: {
                uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`
            },
            position: {
                character: ctx.position.char! - 1,
                line: ctx.position.line - 1
            }
        },
        context: {
            includeDeclaration: true
        }
    } as any, ctx, ctx.filePath)

    return fetch(`/.api/xlang/textDocument/references`, { method: 'POST', body: JSON.stringify(body) })
        .then(resp => resp.json())
        .then(json => {
            if (!json ||
                !json[1] ||
                !json[1].result) {
                return []
            }

            const result = json[1].result
            for (const ref of result) {
                const parsed = new URL(ref.uri)
                ref.repoURI = parsed.hostname + parsed.pathname
            }
            return result
        })
}, makeRepoURI)

interface XReferencesParams extends AbsoluteRepoFile {
    query: string
    hints: any
    limit: number
}

export const fetchXreferences = memoizedFetch((ctx: XReferencesParams): Promise<Reference[]> => {
    const ext = getPathExtension(ctx.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve([])
    }

    const body = wrapLSP({
        method: 'workspace/xreferences',
        params: {
            hints: ctx.hints,
            query: ctx.query,
            limit: ctx.limit
        }
    }, { repoPath: ctx.repoPath, commitID: ctx.commitID }, ctx.filePath)

    return fetch(`/.api/xlang/workspace/xreferences`, { method: 'POST', body: JSON.stringify(body) })
        .then(resp => resp.json())
        .then(json => {
            if (!json ||
                !json[1] ||
                !json[1].result) {
                return []
            }

            return json[1].result.map(res => {
                const ref = res.reference
                const parsed = new URL(ref.uri)
                ref.repoURI = parsed.hostname + parsed.pathname
                return ref
            })
        })
}, ctx => makeRepoURI(ctx) + '___' + ctx.query + '___' + ctx.limit)
