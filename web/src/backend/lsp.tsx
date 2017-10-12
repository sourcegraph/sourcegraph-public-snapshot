import { Definition, Hover, Location } from 'vscode-languageserver-types'
import { AbsoluteRepo, AbsoluteRepoFile, AbsoluteRepoFilePosition, makeRepoURI, parseRepoURI } from '../repo'
import { getModeFromExtension, getPathExtension, supportedExtensions } from '../util'
import { memoizeAsync } from '../util/memoize'
import { toAbsoluteBlobURL, toPrettyBlobURL } from '../util/url'

interface LSPRequest {
    method: string
    params: any
}

export function isEmptyHover(hover: Hover): boolean {
    return !hover.contents || (Array.isArray(hover.contents) && hover.contents.length === 0)
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
                mode: `${getModeFromExtension(getPathExtension(path))}`,
            },
        },
        {
            id: 1,
            ...req,
        },
        {
            id: 2,
            method: 'shutdown',
        },
        {
            // id not included on 'exit' requests
            method: 'exit',
        },
    ]
}

export const fetchHover = memoizeAsync((pos: AbsoluteRepoFilePosition): Promise<Hover> => {
    const ext = getPathExtension(pos.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve({ contents: [] })
    }

    const body = wrapLSP({
        method: 'textDocument/hover',
        params: {
            textDocument: {
                uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`,
            },
            position: {
                character: pos.position.character! - 1,
                line: pos.position.line - 1,
            },
        },
    }, pos, pos.filePath)

    return fetch(`/.api/xlang/textDocument/hover`, { method: 'POST', body: JSON.stringify(body), headers: { ...window.context.xhrHeaders }, credentials: 'same-origin' })
        .then(resp => resp.json())
        .then(json => {
            if (!json || !json[1] || !json[1].result) {
                return []
            }
            return json[1].result
        })
}, makeRepoURI)

export const fetchDefinition = memoizeAsync((pos: AbsoluteRepoFilePosition): Promise<Definition> => {
    const ext = getPathExtension(pos.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve([])
    }

    const body = wrapLSP({
        method: 'textDocument/definition',
        params: {
            textDocument: {
                uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`,
            },
            position: {
                character: pos.position.character! - 1,
                line: pos.position.line - 1,
            },
        },
    }, pos, pos.filePath)

    return fetch(`/.api/xlang/textDocument/definition`, { method: 'POST', body: JSON.stringify(body), headers: { ...window.context.xhrHeaders }, credentials: 'same-origin' })
        .then(resp => resp.json())
        .then(json => {
            if (!json || !json[1] || !json[1].result) {
                return []
            }
            return json[1].result
        })
}, makeRepoURI)

export function fetchJumpURL(pos: AbsoluteRepoFilePosition): Promise<string | null> {
    return fetchDefinition(pos)
        .then(def => {
            const defArray = Array.isArray(def) ? def : [def]
            def = defArray[0]
            if (!def) {
                return null
            }

            const uri = parseRepoURI(def.uri) as AbsoluteRepoFilePosition
            uri.position = { line: def.range.start.line + 1, character: def.range.start.character + 1 }
            if (uri.repoPath === pos.repoPath && uri.commitID === pos.commitID) {
                // Use pretty rev from the current context for same-repo J2D.
                uri.rev = pos.rev
                return toPrettyBlobURL(uri)
            }
            return toAbsoluteBlobURL(uri)
        })
}

export const fetchXdefinition = memoizeAsync((pos: AbsoluteRepoFilePosition): Promise<{ location: any, symbol: any } | null> => {
    const body = wrapLSP({
        method: 'textDocument/xdefinition',
        params: {
            textDocument: {
                uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`,
            },
            position: {
                character: pos.position.character! - 1,
                line: pos.position.line - 1,
            },
        },
    }, pos, pos.filePath)

    return fetch(`/.api/xlang/textDocument/xdefinition`, { method: 'POST', body: JSON.stringify(body), headers: { ...window.context.xhrHeaders }, credentials: 'same-origin' })
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

export const fetchReferences = memoizeAsync((ctx: AbsoluteRepoFilePosition): Promise<Location[]> => {
    const ext = getPathExtension(ctx.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve([])
    }
    const body = wrapLSP({
        method: 'textDocument/references',
        params: {
            textDocument: {
                uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`,
            },
            position: {
                character: ctx.position.character! - 1,
                line: ctx.position.line - 1,
            },
            context: {
                includeDeclaration: true,
            },
        },
    } as any, ctx, ctx.filePath)

    return fetch(`/.api/xlang/textDocument/references`, { method: 'POST', body: JSON.stringify(body), headers: { ...window.context.xhrHeaders }, credentials: 'same-origin' })
        .then(resp => resp.json())
        .then(json => {
            if (!json || !json[1] || !json[1].result) {
                throw new Error('empty references response')
            }
            return json[1].result
        })
}, makeRepoURI)

interface XReferencesParams extends AbsoluteRepoFile {
    query: string
    hints: any
    limit: number
}

export interface PackageDescriptor {
    name: string
    version?: string
    repoURL?: string
}

/**
 * Represents information about a programming construct that can be used to identify and locate the
 * construct's symbol. The identification does not have to be unique, but it should be as unique as
 * possible. It is up to the language server to define the schema of this object.
 *
 * In contrast to `SymbolInformation`, `SymbolDescriptor` includes more concrete, language-specific,
 * metadata about the symbol.
 */
export interface SymbolDescriptor {
    /**
     * The kind of the symbol as a ts.ScriptElementKind
     */
    kind: string

    /**
     * The name of the symbol as returned from TS
     */
    name: string

    /**
     * The kind of the symbol the symbol is contained in, as a ts.ScriptElementKind.
     * Is an empty string if the symbol has no container.
     */
    containerKind: string

    /**
     * The name of the symbol the symbol is contained in, as returned from TS.
     * Is an empty string if the symbol has no container.
     */
    containerName: string

    /**
     * The file path of the file where the symbol is defined in, relative to the workspace rootPath.
     */
    filePath: string

    /**
     * A PackageDescriptor describing the package this symbol belongs to.
     * Is `undefined` if the symbol does not belong to a package.
     */
    package?: PackageDescriptor
}

/**
 * Represents information about a reference to programming constructs like variables, classes,
 * interfaces, etc.
 */
export interface ReferenceInformation {
    /**
     * The location in the workspace where the `symbol` is referenced.
     */
    reference: Location

    /**
     * Metadata about the symbol that can be used to identify or locate its definition.
     */
    symbol: SymbolDescriptor
}

export const fetchXreferences = memoizeAsync((ctx: XReferencesParams): Promise<Location[]> => {
    const ext = getPathExtension(ctx.filePath)
    if (!supportedExtensions.has(ext)) {
        return Promise.resolve([])
    }

    const body = wrapLSP({
        method: 'workspace/xreferences',
        params: {
            hints: ctx.hints,
            query: ctx.query,
            limit: ctx.limit,
        },
    }, { repoPath: ctx.repoPath, commitID: ctx.commitID }, ctx.filePath)

    return fetch(`/.api/xlang/workspace/xreferences`, { method: 'POST', body: JSON.stringify(body), headers: { ...window.context.xhrHeaders }, credentials: 'same-origin' })
        .then(resp => resp.json())
        .then(json => {
            if (!json || !json[1] || !json[1].result) {
                throw new Error('empty xreferences responses')
            }
            return json[1].result.map((data: ReferenceInformation) => data.reference)
        })
}, ctx => makeRepoURI(ctx) + '___' + JSON.stringify(ctx.query) + '___' + ctx.limit)
