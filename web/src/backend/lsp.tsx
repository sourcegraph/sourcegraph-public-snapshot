import {
    ReferenceInformation,
    SymbolLocationInformation,
    WorkspaceReferenceParams,
} from 'javascript-typescript-langserver/lib/request-type'
import { Observable } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map, tap } from 'rxjs/operators'
import { Definition, Hover, Location, MarkedString, MarkupContent } from 'vscode-languageserver-types'
import { DidOpenTextDocumentParams, InitializeResult, ServerCapabilities } from 'vscode-languageserver/lib/main'
import { AbsoluteRepo, FileSpec, makeRepoURI, parseRepoURI, PositionSpec } from '../repo'
import { normalizeAjaxError } from '../util/errors'
import { memoizeObservable } from '../util/memoize'
import { toAbsoluteBlobURL, toPrettyBlobURL } from '../util/url'
import { ModeSpec } from './features'

/**
 * Contains the fields necessary to route the request to the correct logical LSP server process and construct the
 * correct initialization request.
 */
export interface LSPSelector extends AbsoluteRepo, ModeSpec {}

/**
 * Contains the fields necessary to construct an LSP TextDocumentPositionParams value.
 */
export interface LSPTextDocumentPositionParams extends LSPSelector, PositionSpec, FileSpec {}

interface LSPRequest {
    method: string
    params?: any
}
export const isEmptyHover = (hover: Hover | null): boolean =>
    !hover ||
    !hover.contents ||
    (Array.isArray(hover.contents) && hover.contents.length === 0) ||
    (MarkupContent.is(hover.contents) && !hover.contents.value)

/** Returns the first MarkedString element from the hover, or undefined if it has none. */
export function firstMarkedString(hover: Hover): MarkedString | undefined {
    if (typeof hover.contents === 'string') {
        return hover.contents
    } else if (Array.isArray(hover.contents)) {
        return hover.contents[0]
    }
    return hover.contents.value
}

const wrapLSPRequests = (ctx: LSPSelector, requests: LSPRequest[]): any[] => [
    {
        id: 0,
        method: 'initialize',
        params: {
            rootUri: `git://${ctx.repoPath}?${ctx.commitID}`,
            mode: ctx.mode,
        },
    },
    ...requests.map((req, i) => ({ id: i + 1, ...req })),
    {
        id: 1 + requests.length,
        method: 'shutdown',
    },
    {
        // id not included on notifications
        method: 'exit',
    },
]

/** LSP proxy error code for unsupported modes */
export const EMODENOTFOUND = -32000

const sendLSPRequests = (ctx: LSPSelector, ...requests: LSPRequest[]): Observable<any> =>
    ajax({
        method: 'POST',
        url: `/.api/xlang/${requests.map(({ method }) => method).join(',')}`,
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(wrapLSPRequests(ctx, requests)),
    }).pipe(
        // Workaround for https://github.com/ReactiveX/rxjs/issues/3606
        tap(response => {
            if (response.status === 0) {
                throw Object.assign(new Error('Ajax status 0'), response)
            }
        }),
        catchError<AjaxResponse, never>(err => {
            normalizeAjaxError(err)
            throw err
        }),
        map(({ response }) => response),
        map(results => {
            for (const result of results) {
                if (result && result.error) {
                    throw Object.assign(new Error(result.error.message), result.error)
                }
            }

            return results.map((result: any) => result && result.result)
        })
    )

/**
 * Sends an LSP request to the xlang API.
 * If an error is returned, the Promise is rejected with the error from the response.
 *
 * @param req The LSP request to send
 * @param ctx Repo and revision
 * @param mode the LSP mode identifying the language server to use
 * @param path File path for determining the mode
 * @return The result of the method call
 */
const sendLSPRequest = (req: LSPRequest, ctx: LSPSelector): Observable<any> =>
    sendLSPRequests(ctx, req).pipe(map(results => results[1]))

/**
 * Query the server's capabilities. The filePath must be any valid file path, to avoid causing the server to
 * complain about the file not existing, even though the file itself is only used in a noop response.
 */
export const fetchServerCapabilities = memoizeObservable(
    (pos: LSPSelector & { filePath: string }): Observable<ServerCapabilities> =>
        sendLSPRequests(pos, {
            method: 'textDocument/didOpen',
            params: {
                textDocument: {
                    uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`,
                },
            } as DidOpenTextDocumentParams,
        }).pipe(map(results => (results[0] as InitializeResult).capabilities)),
    cacheKey
)

/**
 * Fixes a response to textDocument/hover that is invalid because either
 * `range` or `contents` are `null`.
 *
 * See the spec:
 *
 * https://microsoft.github.io/language-server-protocol/specification#textDocument_hover
 *
 * @param response The LSP response to fix (will be mutated)
 */
const normalizeHoverResponse = (hoverResponse: any): void => {
    // rls for Rust sometimes responds with `range: null`.
    // https://github.com/sourcegraph/sourcegraph/issues/11880
    if (hoverResponse && !hoverResponse.range) {
        hoverResponse.range = undefined
    }

    // clangd for C/C++ sometimes responds with `contents: null`.
    // https://github.com/sourcegraph/sourcegraph/issues/11880#issuecomment-396650342
    if (hoverResponse && !hoverResponse.contents) {
        hoverResponse.contents = []
    }
}

/** Callers should use features.getHover instead. */
export const fetchHover = memoizeObservable(
    (pos: LSPTextDocumentPositionParams): Observable<Hover | null> =>
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
            pos
        ).pipe(
            tap(hover => {
                normalizeHoverResponse(hover)
                // Do some shallow validation on response, e.g. to catch https://github.com/sourcegraph/sourcegraph/issues/11711
                if (hover !== null && !Hover.is(hover)) {
                    throw Object.assign(new Error('Invalid hover response from language server'), { hover })
                }
            })
        ),
    cacheKey
)

/** Callers should use features.getDefinition instead. */
export const fetchDefinition = memoizeObservable(
    (options: LSPTextDocumentPositionParams): Observable<Definition> =>
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
            options
        ),
    cacheKey
)

/** Callers should use features.getJumpURL instead. */
export function fetchJumpURL(options: LSPTextDocumentPositionParams): Observable<string | null> {
    return fetchDefinition(options).pipe(
        map(def => {
            const defArray = Array.isArray(def) ? def : [def]
            def = defArray[0]
            if (!def) {
                return null
            }

            const uri = parseRepoURI(def.uri) as LSPTextDocumentPositionParams
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

/** Callers should use features.getXdefinition instead. */
export const fetchXdefinition = memoizeObservable(
    (options: LSPTextDocumentPositionParams): Observable<SymbolLocationInformation | undefined> =>
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
            options
        ).pipe(map((result: SymbolLocationInformation[]) => result[0])),
    cacheKey
)

export interface LSPReferencesParams {
    includeDeclaration?: boolean
}

/** Callers should use features.getReferences instead. */
export const fetchReferences = memoizeObservable(
    (options: LSPTextDocumentPositionParams & LSPReferencesParams): Observable<Location[]> =>
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
                        includeDeclaration: options.includeDeclaration !== false, // undefined means true
                    },
                },
            },
            options
        ),
    cacheKey
)

/** Callers should use features.getImplementation instead. */
export const fetchImplementation = memoizeObservable(
    (options: LSPTextDocumentPositionParams): Observable<Location[]> =>
        sendLSPRequest(
            {
                method: 'textDocument/implementation',
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
            options
        ),
    cacheKey
)

export interface XReferenceOptions extends WorkspaceReferenceParams {
    /**
     * This is not in the spec, but Go and possibly others support it
     * https://github.com/sourcegraph/go-langserver/blob/885ad3639de0e1e6c230db5395ea0f682534b458/pkg/lspext/lspext.go#L32
     */
    limit: number
}

/** Callers should use features.getXreferences instead. */
export const fetchXreferences = memoizeObservable(
    (options: XReferenceOptions & LSPSelector): Observable<Location[]> =>
        sendLSPRequest(
            {
                method: 'workspace/xreferences',
                params: {
                    hints: options.hints,
                    query: options.query,
                    limit: options.limit,
                },
            },
            {
                repoPath: options.repoPath,
                rev: options.rev,
                commitID: options.commitID,
                mode: options.mode,
            }
        ).pipe(map((refInfos: ReferenceInformation[]) => refInfos.map(refInfo => refInfo.reference))),
    options => cacheKey(options) + ':' + JSON.stringify([options.query, options.hints, options.limit])
)

function cacheKey(sel: LSPSelector): string {
    return `${makeRepoURI(sel)}:mode=${sel.mode}`
}
