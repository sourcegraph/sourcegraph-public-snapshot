import { from, Observable, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { DocumentSelector } from 'sourcegraph'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtLanguageFeaturesAPI } from '../../extension/api/languageFeatures'
import { ReferenceParams, TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { Definition, Hover, Location } from '../../protocol/plainTypes'
import { ProvideTextDocumentHoverSignature } from '../providers/hover'
import { ProvideTextDocumentLocationSignature, TextDocumentReferencesProviderRegistry } from '../providers/location'
import { FeatureProviderRegistry } from '../providers/registry'
import { SubscriptionMap } from './common'

/** @internal */
export interface ClientLanguageFeaturesAPI {
    $unregister(id: number): void
    $registerHoverProvider(id: number, selector: DocumentSelector): void
    $registerDefinitionProvider(id: number, selector: DocumentSelector): void
    $registerTypeDefinitionProvider(id: number, selector: DocumentSelector): void
    $registerImplementationProvider(id: number, selector: DocumentSelector): void
    $registerReferenceProvider(id: number, selector: DocumentSelector): void
}

/** @internal */
export class ClientLanguageFeatures implements ClientLanguageFeaturesAPI {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtLanguageFeaturesAPI

    constructor(
        connection: Connection,
        private hoverRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentHoverSignature
        >,
        private definitionRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentLocationSignature
        >,
        private typeDefinitionRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentLocationSignature
        >,
        private implementationRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentLocationSignature
        >,
        private referencesRegistry: TextDocumentReferencesProviderRegistry
    ) {
        this.subscriptions.add(this.registrations)

        this.proxy = createProxyAndHandleRequests('languageFeatures', connection, this)
    }

    public $unregister(id: number): void {
        this.registrations.remove(id)
    }

    public $registerHoverProvider(id: number, selector: DocumentSelector): void {
        this.registrations.add(
            id,
            this.hoverRegistry.registerProvider(
                { documentSelector: selector },
                (params: TextDocumentPositionParams): Observable<Hover | null | undefined> =>
                    from(this.proxy.$provideHover(id, params.textDocument.uri, params.position))
            )
        )
    }

    public $registerDefinitionProvider(id: number, selector: DocumentSelector): void {
        this.registrations.add(
            id,
            this.definitionRegistry.registerProvider(
                { documentSelector: selector },
                (params: TextDocumentPositionParams): Observable<Definition> =>
                    from(this.proxy.$provideDefinition(id, params.textDocument.uri, params.position)).pipe(
                        map(result => result || [])
                    )
            )
        )
    }

    public $registerTypeDefinitionProvider(id: number, selector: DocumentSelector): void {
        this.registrations.add(
            id,
            this.typeDefinitionRegistry.registerProvider(
                { documentSelector: selector },
                (params: TextDocumentPositionParams): Observable<Definition> =>
                    from(this.proxy.$provideTypeDefinition(id, params.textDocument.uri, params.position)).pipe(
                        map(result => result || [])
                    )
            )
        )
    }

    public $registerImplementationProvider(id: number, selector: DocumentSelector): void {
        this.registrations.add(
            id,
            this.implementationRegistry.registerProvider(
                { documentSelector: selector },
                (params: TextDocumentPositionParams): Observable<Definition> =>
                    from(this.proxy.$provideImplementation(id, params.textDocument.uri, params.position)).pipe(
                        map(result => result || [])
                    )
            )
        )
    }

    public $registerReferenceProvider(id: number, selector: DocumentSelector): void {
        this.registrations.add(
            id,
            this.referencesRegistry.registerProvider(
                { documentSelector: selector },
                (params: ReferenceParams): Observable<Location[]> =>
                    from(
                        this.proxy.$provideReferences(id, params.textDocument.uri, params.position, params.context)
                    ).pipe(map(result => result || []))
            )
        )
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
