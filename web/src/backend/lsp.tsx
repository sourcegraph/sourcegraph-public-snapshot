import { Definition, Hover, Location } from 'vscode-languageserver-types'
import { AbsoluteRepo, AbsoluteRepoFile, AbsoluteRepoFilePosition, makeRepoURI, parseRepoURI } from '../repo'
import { getModeFromExtension, getPathExtension, supportedExtensions } from '../util'
import { memoizeAsync } from '../util/memoize'
import { toAbsoluteBlobURL, toPrettyBlobURL } from '../util/url'

interface LSPRequest {
    method: string
    params: any
}

export const isEmptyHover = (hover: any): boolean =>
    !hover || !hover.contents || (Array.isArray(hover.contents) && hover.contents.length === 0)

const getHeaders = (): Headers => {
    const headers = new Headers()
    for (const [key, value] of Object.entries(window.context.xhrHeaders)) {
        headers.set(key, value)
    }
    return headers
}

const wrapLSPRequest = (req: LSPRequest, ctx: AbsoluteRepo, path: string): any[] => [
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
        // id not included on notifications
        method: 'exit',
    },
]

/** LSP proxy error code for unsupported modes */
export const EMODENOTFOUND = -32000

/**
 * A static list of what is supported on Sourcegraph.com. However, on-prem instances may support
 * less so we prevent future requests after the first mode not found response.
 */
const unsupportedExtensions = new Set<string>()

const isSupported = (path: string): boolean => {
    const ext = getPathExtension(path)
    return supportedExtensions.has(ext) && !unsupportedExtensions.has(ext)
}

/**
 * Sends an LSP request to the xlang API.
 * If an error is returned, the Promise is rejected with the error from the response.
 *
 * @param req The LSP request to send
 * @param ctx Repo and revision
 * @param path File path for determining the mode
 * @return The result of the method call
 */
const sendLSPRequest = (req: LSPRequest, ctx: AbsoluteRepo, path: string): Promise<any> => {
    if (!isSupported(path)) {
        return Promise.reject(Object.assign(new Error('Language not supported'), { code: EMODENOTFOUND }))
    }
    return fetch(`/.api/xlang/${req.method}`, {
        method: 'POST',
        body: JSON.stringify(wrapLSPRequest(req, ctx, path)),
        headers: getHeaders(),
        credentials: 'same-origin',
    })
        .then(response => response.json())
        .then(results => {
            for (const result of results) {
                if (result && result.error) {
                    if (result.error.code === EMODENOTFOUND) {
                        unsupportedExtensions.add(getPathExtension(path))
                    }
                    throw Object.assign(new Error(result.error.message), result.error)
                }
            }
            return results[1].result
        })
}

export const fetchHover = memoizeAsync(
    (pos: AbsoluteRepoFilePosition): Promise<Hover> =>
        sendLSPRequest(
            {
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
            },
            pos,
            pos.filePath
        ),
    makeRepoURI
)

export const fetchDefinition = memoizeAsync(
    (pos: AbsoluteRepoFilePosition): Promise<Definition> =>
        sendLSPRequest(
            {
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
            },
            pos,
            pos.filePath
        ),
    makeRepoURI
)

export function fetchJumpURL(pos: AbsoluteRepoFilePosition): Promise<string | null> {
    return fetchDefinition(pos).then(def => {
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

type XDefinitionResponse = { location: any; symbol: any } | null
export const fetchXdefinition = memoizeAsync(
    (pos: AbsoluteRepoFilePosition): Promise<XDefinitionResponse> =>
        sendLSPRequest(
            {
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
            },
            pos,
            pos.filePath
        ),
    makeRepoURI
)

export const fetchReferences = memoizeAsync(
    (ctx: AbsoluteRepoFilePosition): Promise<Location[]> =>
        sendLSPRequest(
            {
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
            } as any,
            ctx,
            ctx.filePath
        ),
    makeRepoURI
)

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

export const fetchXreferences = memoizeAsync(
    (ctx: XReferencesParams): Promise<Location[]> =>
        sendLSPRequest(
            {
                method: 'workspace/xreferences',
                params: {
                    hints: ctx.hints,
                    query: ctx.query,
                    limit: ctx.limit,
                },
            },
            { repoPath: ctx.repoPath, commitID: ctx.commitID },
            ctx.filePath
        ).then((refInfos: ReferenceInformation[]) => refInfos.map(refInfo => refInfo.reference)),
    ctx => makeRepoURI(ctx) + '___' + ctx.query + '___' + ctx.limit
)
