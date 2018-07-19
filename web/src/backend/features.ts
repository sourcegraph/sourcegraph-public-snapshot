import { HoverMerged } from 'cxp/lib/types/hover'
import { SymbolLocationInformation } from 'javascript-typescript-langserver/lib/request-type'
import { compact, flatten } from 'lodash'
import { forkJoin, Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { Definition, Location } from 'vscode-languageserver-types'
import { CXPControllerProps, USE_PLATFORM } from '../cxp/CXPEnvironment'
import { ConfiguredExtension, ExtensionSettings } from '../extensions/extension'
import { AbsoluteRepo, AbsoluteRepoFile, parseRepoURI } from '../repo'
import { toAbsoluteBlobURL, toPrettyBlobURL } from '../util/url'
import {
    fetchDecorations,
    fetchDefinition,
    fetchHover,
    fetchImplementation,
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
export type Extensions = ConfiguredExtension[]

/** Extended by React prop types that carry extensions. */
export interface ExtensionsProps {
    /** The enabled extensions. */
    extensions: Extensions
}

/**
 * Contains the props needed by this file's getXyz functions to communicate with extensions (using the new CXP
 * implementation) or with language servers (using the old LSP HTTP POST implementation).
 */
interface ExtensionsAndCXPControllerProps extends ExtensionsProps, CXPControllerProps {}

/** Extended by React prop types for components that need to signal a change to extensions. */
export interface ExtensionsChangeProps {
    onExtensionsChange: (enabledExtensions: Extensions) => void
}

export { HoverMerged } // reexport to avoid needing to change all import sites - TODO(sqs): actually go change all them

/**
 * Fetches hover information for the given location.
 *
 * @param ctx the location
 * @return hover for the location
 */
export function getHover(
    ctx: LSPTextDocumentPositionParams,
    { extensions, cxpController }: ExtensionsAndCXPControllerProps
): Observable<HoverMerged | null> {
    if (USE_PLATFORM) {
        return cxpController.registries.textDocumentHover.getHover({
            textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
            position: {
                character: ctx.position.character - 1,
                line: ctx.position.line - 1,
            },
        })
    }
    return forkJoin(getModes(ctx, extensions).map(({ mode, settings }) => fetchHover({ ...ctx, mode, settings }))).pipe(
        map(HoverMerged.from)
    )
}

/**
 * Fetches definitions (in the same repository) for the given location.
 *
 * @param ctx the location
 * @return definitions of the symbol at the location
 */
export function getDefinition(
    ctx: LSPTextDocumentPositionParams,
    { extensions, cxpController }: ExtensionsAndCXPControllerProps
): Observable<Definition> {
    if (USE_PLATFORM) {
        return cxpController.registries.textDocumentDefinition.getLocation({
            textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
            position: {
                character: ctx.position.character - 1,
                line: ctx.position.line - 1,
            },
        })
    }
    return forkJoin(
        getModes(ctx, extensions).map(({ mode, settings }) => fetchDefinition({ ...ctx, mode, settings }))
    ).pipe(map(results => flatten(compact(results))))
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
export function getJumpURL(
    ctx: LSPTextDocumentPositionParams,
    extensions: ExtensionsAndCXPControllerProps
): Observable<string | null> {
    return getDefinition(ctx, extensions).pipe(
        map(def => {
            const defArray = Array.isArray(def) ? def : [def]
            def = defArray[0]
            if (!def) {
                return null
            }

            const uri = parseRepoURI(def.uri) as LSPTextDocumentPositionParams
            uri.position = { line: def.range.start.line + 1, character: def.range.start.character + 1 }
            if (uri.repoPath === ctx.repoPath && uri.commitID === ctx.commitID) {
                // Use pretty rev from the current context for same-repo J2D.
                uri.rev = ctx.rev
                return toPrettyBlobURL(uri)
            }
            return toAbsoluteBlobURL(uri)
        })
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
    return forkJoin(
        getModes(ctx, extensions).map(({ mode, settings }) => fetchXdefinition({ ...ctx, mode, settings }))
    ).pipe(map(results => results.find(v => !!v)))
}

/**
 * Wrap the value in an array. Unlike Lodash's castArray, it maps null to [] (not [null]).
 */
function castArray<T>(value: null | T | T[]): T[] {
    if (value === null) {
        return []
    }
    if (!Array.isArray(value)) {
        return [value]
    }
    return value
}

/**
 * Fetches references (in the same repository) to the symbol at the given location.
 *
 * @param ctx the location
 * @return references to the symbol at the location
 */
export function getReferences(
    ctx: LSPTextDocumentPositionParams & LSPReferencesParams,
    { extensions, cxpController }: ExtensionsAndCXPControllerProps
): Observable<Location[]> {
    if (USE_PLATFORM) {
        return cxpController.registries.textDocumentReferences
            .getLocation({
                textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
                position: {
                    character: ctx.position.character - 1,
                    line: ctx.position.line - 1,
                },
                context: {
                    includeDeclaration: ctx.includeDeclaration !== false, // undefined means true
                },
            })
            .pipe(map(castArray))
    }
    return forkJoin(
        getModes(ctx, extensions).map(({ mode, settings }) => fetchReferences({ ...ctx, mode, settings }))
    ).pipe(map(results => flatten(results)))
}

/**
 * Fetches implementations (in the same repository) of the symbol at the given location.
 *
 * @param ctx the location
 * @return implementations of the symbol at the location
 */
export function getImplementations(
    ctx: LSPTextDocumentPositionParams,
    { extensions, cxpController }: ExtensionsAndCXPControllerProps
): Observable<Location[]> {
    if (USE_PLATFORM) {
        return cxpController.registries.textDocumentImplementation
            .getLocation({
                textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
                position: {
                    character: ctx.position.character - 1,
                    line: ctx.position.line - 1,
                },
            })
            .pipe(map(castArray))
    }
    return forkJoin(
        getModes(ctx, extensions).map(({ mode, settings }) => fetchImplementation({ ...ctx, mode, settings }))
    ).pipe(map(results => flatten(results)))
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
    return forkJoin(
        getModes(ctx, extensions).map(({ mode, settings }) => fetchXreferences({ ...ctx, mode, settings }))
    ).pipe(map(results => flatten(results)))
}

/**
 * Fetches decorations for the given file.
 *
 * @param ctx the file
 * @return decorations
 */
export function getDecorations(
    ctx: AbsoluteRepoFile & LSPSelector,
    { extensions, cxpController }: ExtensionsAndCXPControllerProps
): Observable<TextDocumentDecoration[] | null> {
    if (USE_PLATFORM) {
        return cxpController.registries.textDocumentDecoration.getDecorations({
            textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
        })
    }
    return forkJoin(
        getModes(ctx, extensions).map(({ mode, settings }) =>
            fetchDecorations({ ...ctx, mode, settings }).pipe(
                map(results => (results === null ? [] : compact(results))),
                catchError(error => {
                    if (!isMethodNotFoundError(error)) {
                        console.error(error)
                    }
                    return [[]]
                })
            )
        )
    ).pipe(map(results => flatten(results)))
}

/** Computes the set of LSP modes to use. */
function getModes(ctx: ModeSpec, extensions: Extensions): { mode: string; settings: ExtensionSettings | null }[] {
    // Using the old behavior (use the language server identified by the mode) when there are no extensions set is
    // the most reliable way of supporting both platform-enabled and non-platform-enabled users. A user would only
    // have a non-empty extensions list if they were platform-enabled. The one weird thing is that if a
    // platform-enabled user disables all extensions, then they will get the default behavior (of using the mode's
    // language server); that is acceptable.
    if (extensions.length === 0) {
        return [{ mode: ctx.mode, settings: null }]
    }
    return extensions.map(({ extensionID, settings }) => ({ mode: extensionID, settings }))
}
