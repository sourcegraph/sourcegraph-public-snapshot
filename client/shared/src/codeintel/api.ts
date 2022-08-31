import {
    Contributions,
    Evaluated,
    fromHoverMerged,
    HoverMerged,
    TextDocumentIdentifier,
    TextDocumentPositionParameters,
} from '@sourcegraph/client-api'
import { parse } from '@sourcegraph/template-parser'
import * as sourcegraph from './legacy-extensions/api'
import * as sglegacy from 'sourcegraph'
import * as clientType from '@sourcegraph/extension-api-types'
import { languageSpecs } from './legacy-extensions/language-specs/languages'
import { LanguageSpec } from './legacy-extensions/language-specs/spec'
import { DocumentSelector, TextDocument } from 'sourcegraph'
import { match } from '../api/client/types/textDocument'
import { getModeFromPath } from '../languages'
import { parseRepoURI } from '../util/url'
import { createProviders, emptySourcegraphProviders, SourcegraphProviders } from './legacy-extensions/providers'
import { RedactingLogger } from './legacy-extensions/logging'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { toPosition } from '../api/extension/api/types'
import { castArray } from 'lodash'
import { isDefined } from '@sourcegraph/common/src/types'
import { ContributionOptions } from '../api/extension/extensionHostApi'
import { evaluateContributions } from '../api/extension/api/contribution'

export type QueryGraphQLFn<T> = () => Promise<T>

export interface CodeIntelAPI {
    hasReferenceProvidersForDocument(textParameters: TextDocumentPositionParameters): Observable<boolean>
    getDefinition(textParameters: TextDocumentPositionParameters): Observable<clientType.Location[]>
    getReferences(
        textParameters: TextDocumentPositionParameters,
        context: sourcegraph.ReferenceContext
    ): Observable<clientType.Location[]>
    getImplementations(parameters: TextDocumentPositionParameters): Observable<clientType.Location[]>
    getContributions(contributionOptions?: ContributionOptions): Evaluated<Contributions>
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
            map(result =>
                castArray(result)
                    .filter(isDefined)
                    .map(location => ({ ...location, uri: location.uri.toString() }))
            )
        )
    }

    hasReferenceProvidersForDocument(textParameters: TextDocumentPositionParameters): Observable<boolean> {
        return of(true)
    }
    getReferences(
        textParameters: TextDocumentPositionParameters,
        context: sourcegraph.ReferenceContext
    ): Observable<clientType.Location[]> {
        const request = requestFor(textParameters)
        return this.locationResult(
            request.providers.references.provideReferences(request.document, request.position, context)
        )
    }
    getDefinition(textParameters: TextDocumentPositionParameters): Observable<clientType.Location[]> {
        const request = requestFor(textParameters)
        return this.locationResult(request.providers.definition.provideDefinition(request.document, request.position))
    }
    getImplementations(textParameters: TextDocumentPositionParameters): Observable<clientType.Location[]> {
        const request = requestFor(textParameters)
        return this.locationResult(
            request.providers.implementations.provideLocations(request.document, request.position)
        )
    }
    getHover(textParameters: TextDocumentPositionParameters): Observable<HoverMerged> {
        const request = requestFor(textParameters)
        return request.providers.hover
            .provideHover(request.document, request.position)
            .pipe(map(result => fromHoverMerged([result]) || { contents: [] }))
    }
    getDocumentHighlights(textParameters: TextDocumentPositionParameters): Observable<sglegacy.DocumentHighlight[]> {
        const request = requestFor(textParameters)
        return request.providers.documentHighlights.provideDocumentHighlights(request.document, request.position).pipe(
            map(result => {
                console.log({ result })
                return result ? (result as sglegacy.DocumentHighlight[]) : []
            })
        )
    }
    getContributions(contributionOptions?: ContributionOptions<unknown> | undefined): Evaluated<Contributions> {
        if (!contributionOptions) {
            return {}
        }
        console.log(contributionOptions)
        const contributions: Contributions = {
            menus: {
                hover: [{ action: 'findImplementations' }],
            },
            actions: [
                {
                    id: 'findImplementations',
                    title: parse<string>('"Find implementations"'),
                    command: 'open',
                    commandArguments: [
                        parse<string>(
                            "${get(context, 'implementations_LANGID') && get(context, 'panel.url') && sub(get(context, 'panel.url'), 'panelID', 'implementations_LANGID') || 'noop'}"
                        ),
                    ],
                    actionItem: { label: parse<string>('"Find implementations"') },
                },
            ],
        }
        const x = evaluateContributions({}, contributions)
        console.log({ x })
        return x
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
