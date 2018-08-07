import { Observable, throwError } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map, tap } from 'rxjs/operators'
import { Definition, Hover } from 'vscode-languageserver-types'
import { InitializeResult, ServerCapabilities } from 'vscode-languageserver/lib/main'
import { AbsoluteRepo, FileSpec, makeRepoURI, PositionSpec } from '../repo'
import { repoUrlCache, sourcegraphUrl } from '../util/context'
import { memoizeObservable } from '../util/memoize'
import { ErrorLike, normalizeAjaxError } from './errors'
import { getHeaders } from './headers'
import { LSPRequest } from './lsp'

/** JSON-RPC2 error for methods that are not found */
const EMETHODNOTFOUND = -32601

/** Returns whether the LSP message is a "method not found" error. */
function isMethodNotFoundError(val: any): boolean {
    return val && val.code === EMETHODNOTFOUND
}

/**
 * Contains the fields necessary to route the request to the correct logical LSP server process and construct the
 * correct initialization request.
 */
export interface LSPSelector extends AbsoluteRepo {
    /** The LSP mode, which identifies the language server to use. */
    mode: string
}

/**
 * Contains the fields necessary to construct an LSP TextDocumentPositionParams value.
 */
export interface LSPTextDocumentPositionParams extends LSPSelector, PositionSpec, FileSpec {}

type ResponseMessages = { 0: { result: InitializeResult } } & any[]

type ResponseResults = { 0: InitializeResult } & any[]

type ResponseError = ErrorLike & { responses: ResponseMessages }

function httpSendLSPRequests(ctx: LSPSelector, request?: LSPRequest): Observable<ResponseResults> {
    const url = repoUrlCache[ctx.repoPath] || sourcegraphUrl
    if (!url) {
        return throwError(new Error(`unable to send request: no Sourcegraph URL found for repository ${ctx.repoPath}`))
    }

    return ajax({
        method: 'POST',
        url: `${url}/.api/xlang/${request ? request.method : 'initialize'}`,
        headers: getHeaders(),
        crossDomain: true,
        withCredentials: true,
        async: true,
        body: JSON.stringify(
            [
                {
                    id: 0,
                    method: 'initialize',
                    params: {
                        rootUri: `git://${ctx.repoPath}?${ctx.commitID}`,
                        mode: ctx.mode,
                        initializationOptions: {
                            mode: ctx.mode,
                        },
                    },
                },
                request ? { id: 1, ...request } : null,
                { id: 2, method: 'shutdown' },
                { method: 'exit' },
            ].filter(m => m !== null)
        ),
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
        map((results: ResponseMessages) => {
            for (const result of results) {
                if (result && result.error) {
                    throw Object.assign(new Error(result.error.message), result.error, { responses: results })
                }
            }

            return results.map(result => result && result.result) as ResponseResults
        })
    )
}

const sendLSPRequest: (ctx: LSPSelector, request?: LSPRequest) => Observable<any> = (ctx, request) =>
    httpSendLSPRequests(ctx, request).pipe(map(results => results[request ? 1 : 0]))

/**
 * Query the server's capabilities.
 */
export const fetchServerCapabilities = memoizeObservable(
    (pos: LSPSelector & { filePath: string }): Observable<ServerCapabilities> =>
        sendLSPRequest(pos).pipe(map(result => (result as InitializeResult).capabilities)),
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

export const fetchHover = memoizeObservable<LSPTextDocumentPositionParams & LSPSelector, Hover | null>(
    (ctx: LSPTextDocumentPositionParams & LSPSelector): Observable<Hover | null> =>
        sendLSPRequest(ctx, {
            method: 'textDocument/hover',
            params: {
                textDocument: {
                    uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`,
                },
                position: {
                    character: ctx.position.character! - 1,
                    line: ctx.position.line - 1,
                },
            },
        }).pipe(
            catchError((error: ResponseError) => {
                // If the language server doesn't support textDocument/hover and it reported that it doesn't
                // support it, ignore the error.
                if (
                    isMethodNotFoundError(error) &&
                    error.responses &&
                    error.responses[0] &&
                    error.responses[0].result.capabilities &&
                    !error.responses[0].result.capabilities.hoverProvider
                ) {
                    return [null]
                }
                return throwError(error)
            }),
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

export const fetchDefinition = memoizeObservable<LSPTextDocumentPositionParams & LSPSelector, Definition>(
    (ctx: LSPTextDocumentPositionParams & LSPSelector): Observable<Definition> =>
        sendLSPRequest(ctx, {
            method: 'textDocument/definition',
            params: {
                textDocument: {
                    uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`,
                },
                position: {
                    character: ctx.position.character! - 1,
                    line: ctx.position.line - 1,
                },
            },
        }),
    cacheKey
)

function cacheKey(sel: LSPSelector): string {
    return `${makeRepoURI(sel)}:mode=${sel.mode}`
}
