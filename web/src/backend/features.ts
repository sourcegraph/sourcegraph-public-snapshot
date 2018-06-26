import { SymbolLocationInformation } from 'javascript-typescript-langserver/lib/request-type'
import { compact, flatten } from 'lodash'
import { forkJoin, Observable, of, throwError } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { Definition, Hover, Location, MarkedString, MarkupContent, Range } from 'vscode-languageserver-types'
import { AbsoluteRepo, AbsoluteRepoFile } from '../repo'
import {
    fetchDecorations,
    fetchDefinition,
    fetchHover,
    fetchImplementation,
    fetchJumpURL,
    fetchReferences,
    fetchXdefinition,
    fetchXreferences,
    isMethodNotFoundError,
    LSPReferencesParams,
    LSPSelector,
    LSPTextDocumentPositionParams,
    TextDocumentDecoration,
    XReferenceOptions,
} from './lsp'

/**
 * Specifies an LSP mode.
 */
export interface ModeSpec {
    /** The LSP mode, which identifies the language server to use. */
    mode: string
}

/** The extensions in use. */
export type Extensions = string[]

/** Extended by React prop types that carry extensions. */
export interface ExtensionsProps {
    /** The enabled extensions. */
    extensions: Extensions
}

/** Extended by React prop types for components that need to signal a change to extensions. */
export interface ExtensionsChangeProps {
    onExtensionsChange: (enabledExtensions: Extensions) => void
}

/** An empty list of extensions, used in components that do not yet support extensions yet. */
export const EXTENSIONS_NOT_SUPPORTED: Extensions = []

/** A hover that is merged from multiple Hover results and normalized. */
export type HoverMerged = Pick<Hover, Exclude<keyof Hover, 'contents'>> & {
    /** Also allows MarkupContent[]. */
    contents: (MarkupContent | MarkedString)[]
}

export namespace HoverMerged {
    /** Reports whether the value conforms to the HoverMerged interface. */
    export function is(value: any): value is HoverMerged {
        // Based on Hover.is from vscode-languageserver-types.
        return (
            value !== null &&
            typeof value === 'object' &&
            Array.isArray(value.contents) &&
            (value.contents as any[]).every(c => MarkupContent.is(c) || MarkedString.is(c)) &&
            (value.range === undefined || Range.is(value.range))
        )
    }
}

/**
 * Fetches hover information for the given location.
 *
 * @param ctx the location
 * @return hover for the location
 */
export function getHover(ctx: LSPTextDocumentPositionParams, extensions: Extensions): Observable<HoverMerged | null> {
    return forkJoin(getModes(ctx, extensions).map(mode => fetchHover({ ...ctx, mode }))).pipe(
        map(results => {
            const contents: HoverMerged['contents'] = []
            let range: HoverMerged['range']
            for (const result of results) {
                if (result) {
                    if (Array.isArray(result.contents)) {
                        contents.push(...result.contents)
                    } else {
                        contents.push(result.contents)
                    }
                    if (result.range && !range) {
                        range = result.range
                    }
                }
            }
            return contents.length === 0 ? null : { contents, range }
        })
    )
}

/**
 * Fetches definitions (in the same repository) for the given location.
 *
 * @param ctx the location
 * @return definitions of the symbol at the location
 */
export function getDefinition(ctx: LSPTextDocumentPositionParams, extensions: Extensions): Observable<Definition> {
    return forkJoin(getModes(ctx, extensions).map(mode => fetchDefinition({ ...ctx, mode }))).pipe(
        map(results => flatten(compact(results)))
    )
}

/**
 * Fetches the destination URL for the "Go to definition" action in the hover.
 *
 * Only the first URL is returned, even if there are results from multiple providers or a provider returns
 * multiple results.
 *
 * @param ctx the location containing the token whose definition to jump to
 * @return destination URL
 */
export function getJumpURL(ctx: LSPTextDocumentPositionParams, extensions: Extensions): Observable<string | null> {
    return forkJoin(getModes(ctx, extensions).map(mode => fetchJumpURL({ ...ctx, mode }))).pipe(
        map(results => results.find(v => v !== null) || null)
    )
}

/**
 * Fetches the repository-independent symbol descriptor for the given location.
 *
 * Only the first result is returned, even if there are results from multiple providers.
 *
 * @param ctx the location
 * @return information about the symbol at the location
 */
export function getXdefinition(
    ctx: LSPTextDocumentPositionParams,
    extensions: Extensions
): Observable<SymbolLocationInformation | undefined> {
    return forkJoin(getModes(ctx, extensions).map(mode => fetchXdefinition({ ...ctx, mode }))).pipe(
        map(results => results.find(v => !!v))
    )
}

/**
 * Fetches references (in the same repository) to the symbol at the given location.
 *
 * @param ctx the location
 * @return references to the symbol at the location
 */
export function getReferences(
    ctx: LSPTextDocumentPositionParams & LSPReferencesParams,
    extensions: Extensions
): Observable<Location[]> {
    return forkJoin(getModes(ctx, extensions).map(mode => fetchReferences({ ...ctx, mode }))).pipe(
        map(results => flatten(results))
    )
}

/**
 * Fetches implementations (in the same repository) of the symbol at the given location.
 *
 * @param ctx the location
 * @return implementations of the symbol at the location
 */
export function getImplementations(ctx: LSPTextDocumentPositionParams, extensions: Extensions): Observable<Location[]> {
    return forkJoin(getModes(ctx, extensions).map(mode => fetchImplementation({ ...ctx, mode }))).pipe(
        map(results => flatten(results))
    )
}

/**
 * Fetches references in the repository to the symbol described by the repository-independent symbol descriptor.
 *
 * @param ctx the symbol descriptor and repository to search in
 * @return references to the symbol
 */
export function getXreferences(
    ctx: XReferenceOptions & AbsoluteRepo & LSPSelector,
    extensions: Extensions
): Observable<Location[]> {
    return forkJoin(getModes(ctx, extensions).map(mode => fetchXreferences({ ...ctx, mode }))).pipe(
        map(results => flatten(results))
    )
}

/**
 * Fetches decorations for the given file.
 *
 * @param ctx the file
 * @return decorations
 */
export function getDecorations(
    ctx: AbsoluteRepoFile & LSPSelector,
    extensions: Extensions
): Observable<TextDocumentDecoration[]> {
    if (!window.context.platformEnabled) {
        return of([])
    }
    return forkJoin(
        getModes(ctx, extensions).map(mode =>
            fetchDecorations({ ...ctx, mode }).pipe(
                catchError(error => {
                    if (isMethodNotFoundError(error)) {
                        return [[]]
                    }
                    return throwError(error)
                })
            )
        )
    ).pipe(map(results => flatten(results)))
}

/** Computes the set of LSP modes to use. */
function getModes(ctx: ModeSpec, extensions: Extensions): string[] {
    if (extensions.length === 0 || !window.context.platformEnabled) {
        return [ctx.mode]
    }
    return extensions
}
