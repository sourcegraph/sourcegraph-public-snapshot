import {
    ReferenceInformation,
    SymbolLocationInformation,
    WorkspaceReferenceParams,
} from 'javascript-typescript-langserver/lib/request-type'
import { Observable } from 'rxjs/Observable'
import { ajax } from 'rxjs/observable/dom/ajax'
import { ErrorObservable } from 'rxjs/observable/ErrorObservable'
import { map } from 'rxjs/operators/map'
import { Definition, Hover, Location } from 'vscode-languageserver-types'
import { AbsoluteRepo, AbsoluteRepoFile, AbsoluteRepoFilePosition, makeRepoURI, parseRepoURI } from '../repo'
import { getModeFromExtension, getPathExtension, supportedExtensions } from '../util'
import { memoizeObservable } from '../util/memoize'
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
const sendLSPRequest = (req: LSPRequest, ctx: AbsoluteRepo, path: string): Observable<any> => {
    if (!isSupported(path)) {
        return ErrorObservable.create(Object.assign(new Error('Language not supported'), { code: EMODENOTFOUND }))
    }

    return ajax({
        method: 'POST',
        url: `/.api/xlang/${req.method}`,
        headers: getHeaders(),
        body: JSON.stringify(wrapLSPRequest(req, ctx, path)),
    }).pipe(
        map(({ response }) => response),
        map(results => {
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
    )
}

export const fetchHover = memoizeObservable(
    (pos: AbsoluteRepoFilePosition): Observable<Hover> =>
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

export const fetchDefinition = memoizeObservable(
    (options: AbsoluteRepoFilePosition): Observable<Definition> =>
        sendLSPRequest(
            {
                method: 'textDocument/definition',
                params: {
                    textDocument: {
                        uri: `git://${options.repoPath}?${options.commitID}#${options.filePath}`,
                    },
                    position: {
                        character: options.position.character! - 1,
                        line: options.position.line - 1,
                    },
                },
            },
            options,
            options.filePath
        ),
    makeRepoURI
)

export function fetchJumpURL(options: AbsoluteRepoFilePosition): Observable<string | null> {
    return fetchDefinition(options).pipe(
        map(def => {
            const defArray = Array.isArray(def) ? def : [def]
            def = defArray[0]
            if (!def) {
                return null
            }

            const uri = parseRepoURI(def.uri) as AbsoluteRepoFilePosition
            uri.position = { line: def.range.start.line + 1, character: def.range.start.character + 1 }
            if (uri.repoPath === options.repoPath && uri.commitID === options.commitID) {
                // Use pretty rev from the current context for same-repo J2D.
                uri.rev = options.rev
                return toPrettyBlobURL(uri)
            }
            return toAbsoluteBlobURL(uri)
        })
    )
}

export const fetchXdefinition = memoizeObservable(
    (options: AbsoluteRepoFilePosition): Observable<SymbolLocationInformation | undefined> =>
        sendLSPRequest(
            {
                method: 'textDocument/xdefinition',
                params: {
                    textDocument: {
                        uri: `git://${options.repoPath}?${options.commitID}#${options.filePath}`,
                    },
                    position: {
                        character: options.position.character! - 1,
                        line: options.position.line - 1,
                    },
                },
            },
            options,
            options.filePath
        ).pipe(map((result: SymbolLocationInformation[]) => result[0])),
    makeRepoURI
)

export const fetchReferences = memoizeObservable(
    (options: AbsoluteRepoFilePosition): Observable<Location[]> =>
        sendLSPRequest(
            {
                method: 'textDocument/references',
                params: {
                    textDocument: {
                        uri: `git://${options.repoPath}?${options.commitID}#${options.filePath}`,
                    },
                    position: {
                        character: options.position.character! - 1,
                        line: options.position.line - 1,
                    },
                    context: {
                        includeDeclaration: true,
                    },
                },
            },
            options,
            options.filePath
        ),
    makeRepoURI
)

interface XReferenceOptions extends WorkspaceReferenceParams, AbsoluteRepoFile {
    /**
     * This is not in the spec, but Go and possibly others support it
     * https://github.com/sourcegraph/go-langserver/blob/885ad3639de0e1e6c230db5395ea0f682534b458/pkg/lspext/lspext.go#L32
     */
    limit: number
}

export const fetchXreferences = memoizeObservable(
    (options: XReferenceOptions): Observable<Location[]> =>
        sendLSPRequest(
            {
                method: 'workspace/xreferences',
                params: {
                    hints: options.hints,
                    query: options.query,
                    limit: options.limit,
                },
            },
            { repoPath: options.repoPath, commitID: options.commitID },
            options.filePath
        ).pipe(map((refInfos: ReferenceInformation[]) => refInfos.map(refInfo => refInfo.reference))),
    options => makeRepoURI(options) + '___' + options.query + '___' + options.limit
)
