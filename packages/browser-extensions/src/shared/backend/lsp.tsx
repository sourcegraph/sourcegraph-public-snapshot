import { DiffPart, JumpURLFetcher } from '@sourcegraph/codeintellify'
import { Controller } from '@sourcegraph/extensions-client-common/lib/client/controller'
import { ConfigurationSubject, Settings } from '@sourcegraph/extensions-client-common/lib/settings'
import { from, Observable, of, OperatorFunction, throwError, throwError as error } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map, switchMap, tap } from 'rxjs/operators'
import { HoverMerged } from 'sourcegraph/module/client/types/hover'
import { TextDocumentPositionParams } from 'sourcegraph/module/protocol'
import { Definition, TextDocumentIdentifier } from 'vscode-languageserver-types'
import { InitializeResult, ServerCapabilities } from 'vscode-languageserver/lib/main'
import {
    AbsoluteRepo,
    AbsoluteRepoFile,
    AbsoluteRepoFilePosition,
    AbsoluteRepoLanguageFile,
    FileSpec,
    makeRepoURI,
    parseRepoURI,
    PositionSpec,
    RepoSpec,
    ResolvedRevSpec,
    RevSpec,
} from '../repo'
import { canFetchForURL, DEFAULT_SOURCEGRAPH_URL, getModeFromPath, repoUrlCache, sourcegraphUrl } from '../util/context'
import { memoizeObservable } from '../util/memoize'
import { toAbsoluteBlobURL } from '../util/url'
import { normalizeAjaxError, NoSourcegraphURLError } from './errors'
import { getHeaders } from './headers'

export interface LSPRequest {
    method: string
    params: any
}

/** LSP proxy error code for unsupported modes */
export const EMODENOTFOUND = -32000

/**
 * Modes that are known to not be supported because the server replied with a mode not found error
 */
const unsupportedModes = new Set<string>()

export function isEmptyHover(hover: HoverMerged | null): boolean {
    return !hover || !hover.contents || (Array.isArray(hover.contents) && hover.contents.length === 0)
}

const request = (url: string, method: string, requests: any[]): Observable<AjaxResponse> =>
    getHeaders(url).pipe(
        switchMap(headers =>
            ajax({
                method: 'POST',
                url: `${url}/.api/xlang/${method || ''}`,
                headers,
                crossDomain: true,
                withCredentials: !(headers && headers.authorization),
                body: JSON.stringify(requests),
                async: true,
            })
        )
    )

export function sendLSPHTTPRequests(requests: any[], url: string = sourcegraphUrl): Observable<any[]> {
    const sendTo = (urlsToTry: string[]): Observable<AjaxResponse> => {
        if (urlsToTry.length === 0) {
            return throwError(new NoSourcegraphURLError())
        }

        const urlToTry = urlsToTry[0]
        const urlPathHint = requests[1] && requests[1].method

        return request(urlToTry, urlPathHint, requests).pipe(
            // Workaround for https://github.com/ReactiveX/rxjs/issues/3606
            tap(response => {
                if (response.status === 0) {
                    throw Object.assign(new Error('Ajax status 0'), response)
                }
            }),
            catchError(err => {
                if (urlsToTry.length === 1) {
                    // We don't have any fallbacks left, so throw the most recent error.
                    throw err
                } else {
                    return sendTo(urlsToTry.slice(1))
                }
            }),
            catchError<AjaxResponse, never>(err => {
                normalizeAjaxError(err)
                throw err
            })
        )
    }

    return sendTo([url, DEFAULT_SOURCEGRAPH_URL].filter(url => canFetchForURL(url))).pipe(
        map(({ response }) => response)
    )
}

function wrapLSP(req: LSPRequest, ctx: AbsoluteRepo, path: string): any[] {
    return [
        {
            id: 0,
            method: 'initialize',
            params: {
                rootUri: `git://${ctx.repoPath}?${ctx.commitID}`,
                initializationOptions: { mode: `${getModeFromPath(path)}` },
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

/**
 * Inspects a response from LSP Proxy and throws an exception if the response
 * has an error. This is intended to be used in rxjs: pipe(...throwIfError)
 */
const extractLSPResponse: OperatorFunction<AjaxResponse, any> = source =>
    source.pipe(
        tap(ajaxResponse => {
            // Workaround for https://github.com/ReactiveX/rxjs/issues/3606
            if (ajaxResponse.status === 0) {
                throw Object.assign(new Error('Ajax status 0'), ajaxResponse)
            }
        }),
        catchError(err => {
            normalizeAjaxError(err)
            throw err
        }),
        map<AjaxResponse, any[]>(({ response }) => response),
        tap(lspResponses => {
            for (const lspResponse of lspResponses) {
                if (lspResponse && lspResponse.error) {
                    throw Object.assign(new Error(lspResponse.error.message), lspResponse.error, {
                        responses: lspResponses,
                    })
                }
            }
        }),
        map(lspResponses => lspResponses[1] && lspResponses[1].result)
    )

const fetchHover = memoizeObservable((pos: AbsoluteRepoFilePosition): Observable<HoverMerged | null> => {
    const mode = getModeFromPath(pos.filePath)
    if (!mode || unsupportedModes.has(mode)) {
        return of({ contents: [] })
    }

    const body = wrapLSP(
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
    )

    const url = repoUrlCache[pos.repoPath] || sourcegraphUrl
    if (!url) {
        throw new Error('Error fetching hover: No URL found.')
    }
    if (!canFetchForURL(url)) {
        return of(null)
    }

    return request(url, 'textDocument/hover', body).pipe(extractLSPResponse)
}, makeRepoURI)

const fetchDefinition = memoizeObservable((pos: AbsoluteRepoFilePosition): Observable<Definition> => {
    const mode = getModeFromPath(pos.filePath)
    if (!mode || unsupportedModes.has(mode)) {
        return of([])
    }

    const body = wrapLSP(
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
    )

    const url = repoUrlCache[pos.repoPath] || sourcegraphUrl
    if (!url) {
        throw new Error('Error fetching definition: No URL found.')
    }
    if (!canFetchForURL(url)) {
        return of([])
    }
    return request(url, 'textDocument/definition', body).pipe(extractLSPResponse)
}, makeRepoURI)

export function fetchJumpURL(
    fetchDefinition: SimpleProviderFns['fetchDefinition'],
    pos: AbsoluteRepoFilePosition
): Observable<string | null> {
    return fetchDefinition(pos).pipe(
        map(def => {
            const defArray = Array.isArray(def) ? def : [def]
            def = defArray[0]
            if (!def) {
                return null
            }

            const uri = parseRepoURI(def.uri) as AbsoluteRepoFilePosition
            uri.position = { line: def.range.start.line + 1, character: def.range.start.character + 1 }
            return toAbsoluteBlobURL(uri)
        })
    )
}

export type JumpURLLocation = RepoSpec & RevSpec & ResolvedRevSpec & FileSpec & PositionSpec & { part?: DiffPart }
export function createJumpURLFetcher(
    fetchDefinition: SimpleProviderFns['fetchDefinition'],
    buildURL: (pos: JumpURLLocation) => string
): JumpURLFetcher {
    return ({ line, character, part, commitID, repoPath, ...rest }) =>
        fetchDefinition({ ...rest, commitID, repoPath, position: { line, character } }).pipe(
            map(def => {
                const defArray = Array.isArray(def) ? def : [def]
                def = defArray[0]
                if (!def) {
                    return null
                }

                const uri = parseRepoURI(def.uri)
                return buildURL({
                    repoPath: uri.repoPath,
                    commitID: uri.commitID!, // LSP proxy always includes a commitID in the URI.
                    rev: uri.repoPath === repoPath && uri.commitID === commitID ? rest.rev : uri.rev!, // If the commitID is the same, keep the rev.
                    filePath: uri.filePath!, // There's never going to be a definition without a file.
                    position: {
                        line: def.range.start.line + 1,
                        character: def.range.start.character + 1,
                    },
                    part,
                })
            })
        )
}

const fetchServerCapabilities = (pos: AbsoluteRepoLanguageFile): Observable<ServerCapabilities | undefined> => {
    // Check if mode is known to not be supported
    const mode = getModeFromPath(pos.filePath)
    if (!mode || unsupportedModes.has(mode)) {
        return error(Object.assign(new Error('Language not supported'), { code: EMODENOTFOUND }))
    }
    const url = repoUrlCache[pos.repoPath] || sourcegraphUrl
    if (!url) {
        throw new Error('Error fetching server capabilities. No URL found.')
    }
    if (!canFetchForURL(url)) {
        return of(undefined)
    }
    return request(
        url,
        'initialize',
        [
            {
                id: 0,
                method: 'initialize',
                params: {
                    rootUri: `git://${pos.repoPath}?${pos.commitID}`,
                    initializationOptions: { mode },
                },
            },
            { id: 1, method: 'shutdown' },
            { method: 'exit' },
        ].filter(m => m !== null)
    ).pipe(
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
                    if (result.error.code === EMODENOTFOUND) {
                        unsupportedModes.add(mode)
                    }
                    throw Object.assign(new Error(result.error.message), result.error)
                }
            }

            return results.map((result: any) => result && result.result)
        }),
        map(results => (results[0] as InitializeResult).capabilities)
    )
}

export interface SimpleProviderFns {
    fetchHover: (pos: AbsoluteRepoFilePosition) => Observable<HoverMerged | null>
    fetchDefinition: (pos: AbsoluteRepoFilePosition) => Observable<Definition>
    fetchServerCapabilities: (pos: AbsoluteRepoLanguageFile) => Observable<ServerCapabilities | undefined>
}

export const lspViaAPIXlang: SimpleProviderFns = {
    fetchHover,
    fetchDefinition,
    fetchServerCapabilities,
}

export const toTextDocumentIdentifier = (pos: AbsoluteRepoFile): TextDocumentIdentifier => ({
    uri: `git://${pos.repoPath}?${pos.commitID}#${pos.filePath}`,
})

const toTextDocumentPositionParams = (pos: AbsoluteRepoFilePosition): TextDocumentPositionParams => ({
    textDocument: toTextDocumentIdentifier(pos),
    position: {
        character: pos.position.character! - 1,
        line: pos.position.line - 1,
    },
})

export const createLSPFromExtensions = (
    extensionsController: Controller<ConfigurationSubject, Settings>
): SimpleProviderFns => ({
    // Use from() to suppress rxjs type incompatibilities between different minor versions of rxjs in
    // node_modules/.
    fetchHover: pos =>
        from(extensionsController.registries.textDocumentHover.getHover(toTextDocumentPositionParams(pos))).pipe(
            map(hover => (hover === null ? HoverMerged.from([]) : hover))
        ),
    fetchDefinition: pos =>
        from(
            extensionsController.registries.textDocumentDefinition.getLocation(toTextDocumentPositionParams(pos))
        ) as Observable<Definition>,
    fetchServerCapabilities,
})
