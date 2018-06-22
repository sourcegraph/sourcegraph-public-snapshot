import { SymbolLocationInformation } from 'javascript-typescript-langserver/lib/request-type'
import { Observable, throwError as error } from 'rxjs'
import { Definition, Hover, Location } from 'vscode-languageserver-types'
import { AbsoluteRepo, AbsoluteRepoFile, AbsoluteRepoFilePosition, FileSpec } from '../repo'
import { getModeFromPath } from '../util'
import {
    EMODENOTFOUND,
    fetchDefinition,
    fetchHover,
    fetchImplementation,
    fetchJumpURL,
    fetchReferences,
    fetchXdefinition,
    fetchXreferences,
    LSPReferencesParams,
    XReferenceOptions,
} from './lsp'

/**
 * Specifies an LSP mode.
 */
export interface ModeSpec {
    /** The LSP mode, which identifies the language server to use. */
    mode: string
}

export function withDefaultMode<C extends AbsoluteRepo & FileSpec, R>(
    ctx: C,
    f: (pos: C & ModeSpec) => Observable<R>
): Observable<R> {
    // Check if mode is known to not be supported
    const mode = getModeFromArg(ctx)
    if (!mode) {
        return error(Object.assign(new Error('Language not supported'), { code: EMODENOTFOUND }))
    }
    return f(Object.assign({}, ctx, { mode })) // TS compiler rejects spread until https://github.com/Microsoft/TypeScript/pull/13288 is merged
}

function getModeFromArg(arg: FileSpec | ModeSpec): string | undefined {
    if ('mode' in arg && arg.mode) {
        return arg.mode
    }
    if ('filePath' in arg && arg.filePath) {
        return getModeFromPath(arg.filePath)
    }
    return undefined
}

/**
 * Fetches hover information for the given location.
 *
 * @param ctx the location
 * @return hover for the location
 */
export function getHover(ctx: AbsoluteRepoFilePosition): Observable<Hover | null> {
    return withDefaultMode(ctx, fetchHover)
}

/**
 * Fetches definitions (in the same repository) for the given location.
 *
 * @param ctx the location
 * @return definitions of the symbol at the location
 */
export function getDefinition(ctx: AbsoluteRepoFilePosition): Observable<Definition> {
    return withDefaultMode(ctx, fetchDefinition)
}

/**
 * Fetches the destination URL for the "Go to definition" action in the hover.
 *
 * @param ctx the location containing the token whose definition to jump to
 * @return destination URL
 */
export function getJumpURL(ctx: AbsoluteRepoFilePosition): Observable<string | null> {
    return withDefaultMode(ctx, fetchJumpURL)
}

/**
 * Fetches the repository-independent symbol descriptor for the given location.
 *
 * @param ctx the location
 * @return information about the symbol at the location
 */
export function getXdefinition(ctx: AbsoluteRepoFilePosition): Observable<SymbolLocationInformation | undefined> {
    return withDefaultMode(ctx, fetchXdefinition)
}

/**
 * Fetches references (in the same repository) to the symbol at the given location.
 *
 * @param ctx the location
 * @return references to the symbol at the location
 */
export function getReferences(ctx: AbsoluteRepoFilePosition & LSPReferencesParams): Observable<Location[]> {
    return withDefaultMode(ctx, fetchReferences)
}

/**
 * Fetches implementations (in the same repository) of the symbol at the given location.
 *
 * @param ctx the location
 * @return implementations of the symbol at the location
 */
export function getImplementations(ctx: AbsoluteRepoFilePosition): Observable<Location[]> {
    return withDefaultMode(ctx, fetchImplementation)
}

/**
 * Fetches references in the repository to the symbol described by the repository-independent symbol descriptor.
 *
 * @param ctx the symbol descriptor and repository to search in
 * @return references to the symbol
 */
export function getXreferences(ctx: XReferenceOptions & AbsoluteRepoFile): Observable<Location[]> {
    return withDefaultMode(ctx, fetchXreferences)
}
