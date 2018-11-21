import { Observable } from 'rxjs'
import { ajax, AjaxResponse } from 'rxjs/ajax'
import { catchError, map, tap } from 'rxjs/operators'
import { Location } from '../../../shared/src/api/protocol/plainTypes'
import { normalizeAjaxError } from '../../../shared/src/errors'
import { AbsoluteRepo, FileSpec, makeRepoURI, PositionSpec } from '../repo'
import { memoizeObservable } from '../util/memoize'
import { HoverMerged, ModeSpec } from './features'

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
export const isEmptyHover = (hover: HoverMerged | null): boolean =>
    !hover ||
    !hover.contents ||
    (Array.isArray(hover.contents) && hover.contents.length === 0) ||
    hover.contents.every(c => {
        if (typeof c === 'string') {
            return !c
        }
        return !c.value
    })

export function sendLSPHTTPRequests(requests: any[], urlPathHint?: string): Observable<any> {
    if (!urlPathHint) {
        urlPathHint = requests[1] && requests[1].method
    }
    return ajax({
        method: 'POST',
        url: `/.api/xlang/${urlPathHint || '_'}`,
        headers: {
            ...window.context.xhrHeaders,
            Accept: 'application/json',
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(requests),
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
        map(({ response }) => response)
    )
}

const httpSendLSPRequest = (ctx: LSPSelector, request?: LSPRequest): Observable<any[]> =>
    sendLSPHTTPRequests(
        [
            {
                id: 0,
                method: 'initialize',
                params: {
                    rootUri: `git://${ctx.repoPath}?${ctx.commitID}`,
                    mode: ctx.mode,
                    initializationOptions: { mode: ctx.mode },
                },
            },
            request ? { id: 1, ...request } : null,
            { id: 2, method: 'shutdown' },
            { method: 'exit' },
        ].filter(m => m !== null),
        request ? request.method : 'initialize'
    ).pipe(
        map((results: any[]) => {
            for (const result of results) {
                if (result && result.error) {
                    throw Object.assign(new Error(result.error.message), result.error, { responses: results })
                }
            }
            return results.map(result => result && result.result) as any[]
        })
    )

/**
 * Sends a sequence of LSP requests: initialize, request (the arg), shutdown, exit. The result from the request
 * (the arg) is returned. If the request arg is not given, it is omitted and the result from initialize is
 * returned.
 */
const sendLSPRequest: (ctx: LSPSelector, request?: LSPRequest) => Observable<any> = (ctx, request) =>
    httpSendLSPRequest(ctx, request).pipe(map(results => results[request ? 1 : 0]))

/**
 * @todo Remove after migration to new extension-based cross-repository references implementation.
 */
export interface SymbolLocationInformation {
    /**
     * The location where the symbol is defined, if any
     */
    location?: Location
    /**
     * Metadata about the symbol that can be used to identify or locate its definition.
     */
    symbol: SymbolDescriptor
}

/**
 * Represents information about a programming construct that can be used to identify and locate the
 * construct's symbol. The identification does not have to be unique, but it should be as unique as
 * possible. It is up to the language server to define the schema of this object.
 *
 * In contrast to `SymbolInformation`, `SymbolDescriptor` includes more concrete, language-specific,
 * metadata about the symbol.
 *
 * @todo Remove after migration to new extension-based cross-repository references implementation.
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
    package?: {
        name: string
        version?: string
        repoURL?: string
    }
}

/** Callers should use features.getXdefinition instead. */
export const fetchXdefinition = memoizeObservable(
    (ctx: LSPTextDocumentPositionParams & LSPSelector): Observable<SymbolLocationInformation | undefined> =>
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

interface WorkspaceReferenceParams {
    /**
     * Metadata about the symbol that is being searched for.
     */
    query: Partial<SymbolDescriptor>
    /**
     * Hints provides optional hints about where the language server should look in order to find
     * the symbol (this is an optimization). It is up to the language server to define the schema of
     * this object.
     */
    hints?: {
        dependeePackageName?: string
    }
}

export interface XReferenceOptions extends WorkspaceReferenceParams {
    /**
     * This is not in the spec, but Go and possibly others support it
     * https://github.com/sourcegraph/go-langserver/blob/885ad3639de0e1e6c230db5395ea0f682534b458/pkg/lspext/lspext.go#L32
     */
    limit: number
}

/**
 * Represents information about a reference to programming constructs like variables, classes,
 * interfaces, etc.
 *
 * @todo Remove after migration to new extension-based cross-repository references implementation.
 */
interface ReferenceInformation {
    /**
     * The location in the workspace where the `symbol` is referenced.
     */
    reference: Location
    /**
     * Metadata about the symbol that can be used to identify or locate its definition.
     */
    symbol: SymbolDescriptor
}

/** Callers should use features.getXreferences instead. */
export const fetchXreferences = memoizeObservable(
    (ctx: XReferenceOptions & LSPSelector): Observable<Location[]> =>
        sendLSPRequest(
            {
                repoPath: ctx.repoPath,
                rev: ctx.rev,
                commitID: ctx.commitID,
                mode: ctx.mode,
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

function cacheKey(sel: LSPSelector): string {
    return `${makeRepoURI(sel)}:mode=${sel.mode}`
}
