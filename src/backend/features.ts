import { HoverMerged } from '@sourcegraph/codeintellify/lib/types'
import { SymbolLocationInformation } from 'javascript-typescript-langserver/lib/request-type'
import { flatten } from 'lodash'
import { forkJoin, Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { Definition, Location, TextDocumentDecoration } from 'sourcegraph/module/protocol/plainTypes'
import { ExtensionsControllerProps } from '../extensions/ExtensionsClientCommonContext'
import { AbsoluteRepo, AbsoluteRepoFile, parseRepoURI } from '../repo'
import { toAbsoluteBlobURL, toPrettyBlobURL } from '../util/url'
import {
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

export { HoverMerged } // reexport to avoid needing to change all import sites - TODO(sqs): actually go change all them

/**
 * Fetches hover information for the given location.
 *
 * @param ctx the location
 * @return hover for the location
 */
export function getHover(
    ctx: LSPTextDocumentPositionParams,
    { extensionsController }: ExtensionsControllerProps
): Observable<HoverMerged | null> {
    return extensionsController.registries.textDocumentHover
        .getHover({
            textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
            position: {
                character: ctx.position.character - 1,
                line: ctx.position.line - 1,
            },
        })
        .pipe(map(hover => hover as HoverMerged | null))
}

/**
 * Fetches definitions (in the same repository) for the given location.
 *
 * @param ctx the location
 * @return definitions of the symbol at the location
 */
export function getDefinition(
    ctx: LSPTextDocumentPositionParams,
    { extensionsController }: ExtensionsControllerProps
): Observable<Definition> {
    return extensionsController.registries.textDocumentDefinition.getLocation({
        textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
        position: {
            character: ctx.position.character - 1,
            line: ctx.position.line - 1,
        },
    })
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
    extensions: ExtensionsControllerProps
): Observable<string | null> {
    return getDefinition(ctx, extensions).pipe(
        map(def => {
            const defArray = Array.isArray(def) ? def : [def]
            def = defArray[0]
            if (!def) {
                return null
            }

            const uri = parseRepoURI(def.uri) as LSPTextDocumentPositionParams
            if (def.range) {
                uri.position = { line: def.range.start.line + 1, character: def.range.start.character + 1 }
            }
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
export function getXdefinition(ctx: LSPTextDocumentPositionParams): Observable<SymbolLocationInformation | undefined> {
    return forkJoin(getModes(ctx).map(({ mode }) => fetchXdefinition({ ...ctx, mode }))).pipe(
        map(results => results.find(v => !!v))
    )
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
    { extensionsController }: ExtensionsControllerProps
): Observable<Location[]> {
    return extensionsController.registries.textDocumentReferences
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

/**
 * Fetches implementations (in the same repository) of the symbol at the given location.
 *
 * @param ctx the location
 * @return implementations of the symbol at the location
 */
export function getImplementations(
    ctx: LSPTextDocumentPositionParams,
    { extensionsController }: ExtensionsControllerProps
): Observable<Location[]> {
    return extensionsController.registries.textDocumentImplementation
        .getLocation({
            textDocument: { uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}` },
            position: {
                character: ctx.position.character - 1,
                line: ctx.position.line - 1,
            },
        })
        .pipe(map(castArray))
}

/**
 * Fetches references in the repository to the symbol described by the repository-independent symbol descriptor.
 *
 * @param ctx the symbol descriptor and repository to search in
 * @return references to the symbol
 */
export function getXreferences(ctx: XReferenceOptions & AbsoluteRepo & LSPSelector): Observable<Location[]> {
    return forkJoin(getModes(ctx).map(({ mode }) => fetchXreferences({ ...ctx, mode }))).pipe(
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
    { extensionsController }: ExtensionsControllerProps
): Observable<TextDocumentDecoration[] | null> {
    return extensionsController.registries.textDocumentDecoration.getDecorations({
        uri: `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`,
    })
}

/** Computes the set of LSP modes to use. */
function getModes(ctx: ModeSpec): { mode: string }[] {
    return [{ mode: ctx.mode }]
}
