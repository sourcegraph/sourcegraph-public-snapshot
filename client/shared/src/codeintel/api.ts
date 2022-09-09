import { castArray } from 'lodash'
import { Observable, of } from 'rxjs'
import { defaultIfEmpty, map } from 'rxjs/operators'
import * as sglegacy from 'sourcegraph'
import { DocumentSelector, TextDocument } from 'sourcegraph'

import {
    fromHoverMerged,
    HoverMerged,
    TextDocumentIdentifier,
    TextDocumentPositionParameters,
} from '@sourcegraph/client-api'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
// eslint-disable-next-line no-restricted-imports
import { isDefined } from '@sourcegraph/common/src/types'
import * as clientType from '@sourcegraph/extension-api-types'

import { match } from '../api/client/types/textDocument'
import { FlatExtensionHostAPI } from '../api/contract'
import { proxySubscribable } from '../api/extension/api/common'
import { toPosition } from '../api/extension/api/types'
import { getModeFromPath } from '../languages'
import { parseRepoURI } from '../util/url'

import { CodeIntelContext } from './legacy-extensions/api'
import * as sourcegraph from './legacy-extensions/api'
import { LanguageSpec } from './legacy-extensions/language-specs/language-spec'
import { languageSpecs } from './legacy-extensions/language-specs/languages'
import { RedactingLogger } from './legacy-extensions/logging'
import { createProviders, emptySourcegraphProviders, SourcegraphProviders } from './legacy-extensions/providers'

export type QueryGraphQLFn<T> = () => Promise<T>

export interface CodeIntelAPI {
    hasReferenceProvidersForDocument(textParameters: TextDocumentPositionParameters): Observable<boolean>
    getDefinition(textParameters: TextDocumentPositionParameters): Observable<clientType.Location[]>
    getReferences(
        textParameters: TextDocumentPositionParameters,
        context: sourcegraph.ReferenceContext
    ): Observable<clientType.Location[]>
    getImplementations(parameters: TextDocumentPositionParameters): Observable<clientType.Location[]>
    getHover(textParameters: TextDocumentPositionParameters): Observable<HoverMerged>
    getDocumentHighlights(textParameters: TextDocumentPositionParameters): Observable<sglegacy.DocumentHighlight[]>
}

export function newCodeIntelAPI(context: sourcegraph.CodeIntelContext): CodeIntelAPI {
    sourcegraph.updateCodeIntelContext(context)
    return new DefaultCodeIntelAPI()
}

class DefaultCodeIntelAPI implements CodeIntelAPI {
    private locationResult(
        locations: sourcegraph.ProviderResult<sourcegraph.Definition>
    ): Observable<clientType.Location[]> {
        return locations.pipe(
            defaultIfEmpty(),
            map(result =>
                castArray(result)
                    .filter(isDefined)
                    .map(location => ({ ...location, uri: location.uri.toString() }))
            )
        )
    }

    public hasReferenceProvidersForDocument(textParameters: TextDocumentPositionParameters): Observable<boolean> {
        const document = toTextDocument(textParameters.textDocument)
        const providers = findLanguageMatchingDocument(document)?.providers
        return of(!!providers)
    }
    public getReferences(
        textParameters: TextDocumentPositionParameters,
        context: sourcegraph.ReferenceContext
    ): Observable<clientType.Location[]> {
        const request = requestFor(textParameters)
        return this.locationResult(
            request.providers.references.provideReferences(request.document, request.position, context)
        )
    }
    public getDefinition(textParameters: TextDocumentPositionParameters): Observable<clientType.Location[]> {
        const request = requestFor(textParameters)
        return this.locationResult(request.providers.definition.provideDefinition(request.document, request.position))
    }
    public getImplementations(textParameters: TextDocumentPositionParameters): Observable<clientType.Location[]> {
        const request = requestFor(textParameters)
        return this.locationResult(
            request.providers.implementations.provideLocations(request.document, request.position)
        )
    }
    public getHover(textParameters: TextDocumentPositionParameters): Observable<HoverMerged> {
        const request = requestFor(textParameters)
        return (
            request.providers.hover
                .provideHover(request.document, request.position)
                // We intentionally don't use `defaultIfEmpty()` here because
                // that makes the popover load with an empty docstring.
                .pipe(map(result => fromHoverMerged([result]) || { contents: [] }))
        )
    }
    public getDocumentHighlights(
        textParameters: TextDocumentPositionParameters
    ): Observable<sglegacy.DocumentHighlight[]> {
        const request = requestFor(textParameters)
        return request.providers.documentHighlights.provideDocumentHighlights(request.document, request.position).pipe(
            defaultIfEmpty(),
            map(result => (result ? (result as sglegacy.DocumentHighlight[]) : []))
        )
    }
}

interface LanguageRequest {
    providers: SourcegraphProviders
    document: sourcegraph.TextDocument
    position: sourcegraph.Position
}

function requestFor(textParameters: TextDocumentPositionParameters): LanguageRequest {
    const document = toTextDocument(textParameters.textDocument)
    return {
        document,
        position: toPosition(textParameters.position),
        providers: findLanguageMatchingDocument(document)?.providers || emptySourcegraphProviders,
    }
}

function toTextDocument(textDocument: TextDocumentIdentifier): sourcegraph.TextDocument {
    return {
        uri: textDocument.uri,
        languageId: getModeFromPath(parseRepoURI(textDocument.uri).filePath || ''),
        text: undefined,
    }
}

function findLanguageMatchingDocument(textDocument: TextDocumentIdentifier): Language | undefined {
    const document: Pick<TextDocument, 'uri' | 'languageId'> = toTextDocument(textDocument)
    for (const language of languages) {
        if (match(language.selector, document)) {
            return language
        }
    }
    return undefined
}

interface Language {
    spec: LanguageSpec
    selector: DocumentSelector
    providers: SourcegraphProviders
}
const hasImplementationsField = true
const languages: Language[] = languageSpecs.map(spec => ({
    spec,
    selector: selectorForSpec(spec),
    providers: createProviders(spec, hasImplementationsField, new RedactingLogger(console)),
}))

function selectorForSpec(languageSpec: LanguageSpec): DocumentSelector {
    return [
        { language: languageSpec.languageID },
        ...(languageSpec.verbatimFilenames || []).flatMap(filename => [{ pattern: filename }]),
        ...languageSpec.fileExts.flatMap(extension => [{ pattern: `*.${extension}` }]),
    ]
}

// Replaces codeintel functions from the "old" extension/webworker extension API
// with new implementations of code that lives in this repository. The old
// implementation invoked codeintel functions via webworkers, and the codeintel
// implementation lived in a separate repository
// https://github.com/sourcegraph/code-intel-extensions Ideally, we should
// update all the usages of `comlink.Remote<FlatExtensionHostAPI>` with the new
// `CodeIntelAPI` interfaces, but that would require refactoring a lot of files.
// To minimize the risk of breaking changes caused by the deprecation of
// extensions, we monkey patch the old implementation with new implementations.
// The benefit of monkey patching is that we can optionally disable if for
// customers that choose to enable the legacy extensions.
export function injectNewCodeintel(
    old: FlatExtensionHostAPI,
    codeintelContext: CodeIntelContext
): FlatExtensionHostAPI {
    const codeintel = newCodeIntelAPI(codeintelContext)
    function thenMaybeLoadingResult<T>(promise: Observable<T>): Observable<MaybeLoadingResult<T>> {
        return promise.pipe(
            map(result => {
                const maybeLoadingResult: MaybeLoadingResult<T> = { isLoading: false, result }
                return maybeLoadingResult
            })
        )
    }

    const codeintelOverrides: Pick<
        FlatExtensionHostAPI,
        | 'getHover'
        | 'getDocumentHighlights'
        | 'getReferences'
        | 'getDefinition'
        | 'getLocations'
        | 'hasReferenceProvidersForDocument'
    > = {
        hasReferenceProvidersForDocument(textParameters) {
            return proxySubscribable(codeintel.hasReferenceProvidersForDocument(textParameters))
        },
        getLocations(id, parameters) {
            console.log({ id })
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getImplementations(parameters)))
        },
        getDefinition(parameters) {
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getDefinition(parameters)))
        },
        getReferences(parameters, context) {
            return proxySubscribable(thenMaybeLoadingResult(codeintel.getReferences(parameters, context)))
        },
        getDocumentHighlights: (textParameters: TextDocumentPositionParameters) =>
            proxySubscribable(codeintel.getDocumentHighlights(textParameters)),
        getHover: (textParameters: TextDocumentPositionParameters) =>
            proxySubscribable(thenMaybeLoadingResult(codeintel.getHover(textParameters))),
    }

    return new Proxy(old, {
        get(target, prop) {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-explicit-any
            const codeintelFunction = (codeintelOverrides as any)[prop]
            if (codeintelFunction) {
                return codeintelFunction
            }
            // eslint-disable-next-line prefer-rest-params
            return Reflect.get(target, prop, ...arguments)
        },
    })
}
