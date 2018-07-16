import {
    ReferenceInformation,
    SymbolLocationInformation,
    WorkspaceReferenceParams,
} from 'javascript-typescript-langserver/lib/request-type'
import { Observable, throwError } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map, tap } from 'rxjs/operators'
import { Definition, Hover, Location, MarkupContent, Range } from 'vscode-languageserver-types'
import { InitializeResult, ServerCapabilities } from 'vscode-languageserver/lib/main'
import { ExtensionSettings } from '../extensions/extension'
import { AbsoluteRepo, FileSpec, makeRepoURI, PositionSpec } from '../repo'
import { ErrorLike, normalizeAjaxError } from '../util/errors'
import { memoizeObservable } from '../util/memoize'
import { HoverMerged, ModeSpec } from './features'

/**
 * Contains the fields necessary to route the request to the correct logical LSP server process and construct the
 * correct initialization request.
 */
export interface LSPSelector extends AbsoluteRepo, ModeSpec {}

/** All of the context to send to the backend to initialize the correct server for a request. */
export interface LSPContext extends LSPSelector {
    settings: ExtensionSettings | null
}

/**
 * Contains the fields necessary to construct an LSP TextDocumentPositionParams value.
 */
export interface LSPTextDocumentPositionParams extends LSPSelector, PositionSpec, FileSpec {}

export interface LSPRequest {
    method: string
    params?: any
}
export const isEmptyHover = (hover: HoverMerged | null): boolean =>
    !hover ||
    !hover.contents ||
    (Array.isArray(hover.contents) && hover.contents.length === 0) ||
    (MarkupContent.is(hover.contents) && !hover.contents.value)

/** JSON-RPC2 error for methods that are not found */
const EMETHODNOTFOUND = -32601

/** Returns whether the LSP message is a "method not found" error. */
export function isMethodNotFoundError(val: any): boolean {
    return val && val.code === EMETHODNOTFOUND
}

/** LSP proxy error code for unsupported modes */
export const EMODENOTFOUND = -32000

type ResponseMessages = { 0: { result: InitializeResult } } & any[]

export type ResponseResults = { 0: InitializeResult } & any[]

type ResponseError = ErrorLike & { responses: ResponseMessages }

const httpSendLSPRequest = (ctx: LSPContext, request?: LSPRequest): Observable<ResponseResults> =>
    ajax({
        method: 'POST',
        url: `/.api/xlang/${request ? request.method : 'initialize'}`,
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
        },
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
                            settings: ctx.settings,
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

/**
 * Sends a sequence of LSP requests: initialize, request (the arg), shutdown, exit. The result from the request
 * (the arg) is returned. If the request arg is not given, it is omitted and the result from initialize is
 * returned.
 */
const sendLSPRequest: (ctx: LSPContext, request?: LSPRequest) => Observable<any> = (ctx, request) =>
    httpSendLSPRequest(ctx, request).pipe(map(results => results[request ? 1 : 0]))

/**
 * Query the server's capabilities.
 */
export const fetchServerCapabilities = memoizeObservable(
    (pos: LSPContext & { filePath: string }): Observable<ServerCapabilities> =>
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

/** Callers should use features.getHover instead. */
export const fetchHover = memoizeObservable(
    (ctx: LSPTextDocumentPositionParams & LSPContext): Observable<Hover | null> =>
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

/** Callers should use features.getDefinition instead. */
export const fetchDefinition = memoizeObservable(
    (ctx: LSPTextDocumentPositionParams & LSPContext): Observable<Definition> =>
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

/** Callers should use features.getXdefinition instead. */
export const fetchXdefinition = memoizeObservable(
    (ctx: LSPTextDocumentPositionParams & LSPContext): Observable<SymbolLocationInformation | undefined> =>
        sendLSPRequest(ctx, {
            method: 'textDocument/xdefinition',
            params: {
                textDocument: {
                    uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`,
                },
                position: {
                    character: ctx.position.character! - 1,
                    line: ctx.position.line - 1,
                },
            },
        }).pipe(map((result: SymbolLocationInformation[]) => result[0])),
    cacheKey
)

export interface LSPReferencesParams {
    includeDeclaration?: boolean
}

/** Callers should use features.getReferences instead. */
export const fetchReferences = memoizeObservable(
    (ctx: LSPTextDocumentPositionParams & LSPReferencesParams & LSPContext): Observable<Location[]> =>
        sendLSPRequest(ctx, {
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
                    includeDeclaration: ctx.includeDeclaration !== false, // undefined means true
                },
            },
        }),
    cacheKey
)

/** Callers should use features.getImplementation instead. */
export const fetchImplementation = memoizeObservable(
    (ctx: LSPTextDocumentPositionParams & LSPContext): Observable<Location[]> =>
        sendLSPRequest(ctx, {
            method: 'textDocument/implementation',
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

export interface XReferenceOptions extends WorkspaceReferenceParams {
    /**
     * This is not in the spec, but Go and possibly others support it
     * https://github.com/sourcegraph/go-langserver/blob/885ad3639de0e1e6c230db5395ea0f682534b458/pkg/lspext/lspext.go#L32
     */
    limit: number
}

/** Callers should use features.getXreferences instead. */
export const fetchXreferences = memoizeObservable(
    (ctx: XReferenceOptions & LSPContext): Observable<Location[]> =>
        sendLSPRequest(
            {
                repoPath: ctx.repoPath,
                rev: ctx.rev,
                commitID: ctx.commitID,
                mode: ctx.mode,
                settings: ctx.settings,
            },
            {
                method: 'workspace/xreferences',
                params: {
                    hints: ctx.hints,
                    query: ctx.query,
                    limit: ctx.limit,
                },
            }
        ).pipe(map((refInfos: ReferenceInformation[]) => refInfos.map(refInfo => refInfo.reference))),
    options => cacheKey(options) + ':' + JSON.stringify([options.query, options.hints, options.limit])
)

export interface TextDocumentDecoration {
    after?: DecorationAttachmentRenderOptions
    background?: string
    backgroundColor?: string
    border?: string
    borderColor?: string
    borderWidth?: string
    isWholeLine?: boolean
    range: Range
}

export interface DecorationAttachmentRenderOptions {
    backgroundColor?: string
    color?: string
    contentText?: string
    hoverMessage?: string
    linkURL?: string
}

export const fetchDecorations = memoizeObservable(
    (ctx: LSPContext & FileSpec): Observable<TextDocumentDecoration[] | null> =>
        sendLSPRequest(ctx, {
            method: 'textDocument/decorations',
            params: {
                textDocument: {
                    uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`,
                },
            },
        }),
    cacheKey
)

function cacheKey(sel: LSPContext): string {
    return `${makeRepoURI(sel)}:mode=${sel.mode}:settings=${JSON.stringify(sel.settings)}`
}
