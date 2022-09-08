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
// eslint-disable-next-line no-restricted-imports
import { isDefined } from '@sourcegraph/common/src/types'
import * as clientType from '@sourcegraph/extension-api-types'

import { match } from '../api/client/types/textDocument'
import { toPosition } from '../api/extension/api/types'
import { getModeFromPath } from '../languages'
import { parseRepoURI } from '../util/url'

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
