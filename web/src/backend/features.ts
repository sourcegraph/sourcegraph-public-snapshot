import { SymbolLocationInformation } from 'javascript-typescript-langserver/lib/request-type'
import { Observable } from 'rxjs'
import { Definition, Hover, Location } from 'vscode-languageserver-types'
import { AbsoluteRepoFile } from '../repo'
import {
    fetchDefinition,
    fetchHover,
    fetchImplementation,
    fetchJumpURL,
    fetchReferences,
    fetchXdefinition,
    fetchXreferences,
    LSPReferencesParams,
    LSPSelector,
    LSPTextDocumentPositionParams,
    XReferenceOptions,
} from './lsp'

/**
 * Specifies an LSP mode.
 */
export interface ModeSpec {
    /** The LSP mode, which identifies the language server to use. */
    mode: string
}

/**
 * Fetches hover information for the given location.
 *
 * @param ctx the location
 * @return hover for the location
 */
export function getHover(ctx: LSPTextDocumentPositionParams): Observable<Hover | null> {
    return fetchHover(ctx)
}

/**
 * Fetches definitions (in the same repository) for the given location.
 *
 * @param ctx the location
 * @return definitions of the symbol at the location
 */
export function getDefinition(ctx: LSPTextDocumentPositionParams): Observable<Definition> {
    return fetchDefinition(ctx)
}

/**
 * Fetches the destination URL for the "Go to definition" action in the hover.
 *
 * @param ctx the location containing the token whose definition to jump to
 * @return destination URL
 */
export function getJumpURL(ctx: LSPTextDocumentPositionParams): Observable<string | null> {
    return fetchJumpURL(ctx)
}

/**
 * Fetches the repository-independent symbol descriptor for the given location.
 *
 * @param ctx the location
 * @return information about the symbol at the location
 */
export function getXdefinition(ctx: LSPTextDocumentPositionParams): Observable<SymbolLocationInformation | undefined> {
    return fetchXdefinition(ctx)
}

/**
 * Fetches references (in the same repository) to the symbol at the given location.
 *
 * @param ctx the location
 * @return references to the symbol at the location
 */
export function getReferences(ctx: LSPTextDocumentPositionParams & LSPReferencesParams): Observable<Location[]> {
    return fetchReferences(ctx)
}

/**
 * Fetches implementations (in the same repository) of the symbol at the given location.
 *
 * @param ctx the location
 * @return implementations of the symbol at the location
 */
export function getImplementations(ctx: LSPTextDocumentPositionParams): Observable<Location[]> {
    return fetchImplementation(ctx)
}

/**
 * Fetches references in the repository to the symbol described by the repository-independent symbol descriptor.
 *
 * @param ctx the symbol descriptor and repository to search in
 * @return references to the symbol
 */
export function getXreferences(ctx: XReferenceOptions & AbsoluteRepoFile & LSPSelector): Observable<Location[]> {
    return fetchXreferences(ctx)
}
