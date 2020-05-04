import { Remote, ProxyMarked, proxyMarker } from 'comlink'
import { Hover, Location } from '@sourcegraph/extension-api-types'
import { CompletionList, DocumentSelector, Unsubscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { ReferenceParams, TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { ProvideCompletionItemSignature } from '../services/completion'
import { ProvideTextDocumentHoverSignature } from '../services/hover'
import { TextDocumentLocationProviderIDRegistry, TextDocumentLocationProviderRegistry } from '../services/location'
import { FeatureProviderRegistry } from '../services/registry'
import { registerRemoteProvider } from './common'

/** @internal */
export interface ClientLanguageFeaturesAPI extends ProxyMarked {
    $registerHoverProvider(
        selector: DocumentSelector,
        providerFunction: Remote<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Hover | null | undefined>) & ProxyMarked
        >
    ): Unsubscribable & ProxyMarked
    $registerDefinitionProvider(
        selector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked
    $registerReferenceProvider(
        selector: DocumentSelector,
        providerFunction: Remote<((params: ReferenceParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked

    /**
     * @param idStr The `id` argument in the extension's {@link sourcegraph.languages.registerLocationProvider}
     * call.
     */
    $registerLocationProvider(
        idStr: string,
        selector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked

    $registerCompletionItemProvider(
        selector: DocumentSelector,
        providerFunction: Remote<
            ((params: TextDocumentPositionParams) => ProxySubscribable<CompletionList | null | undefined>) & ProxyMarked
        >
    ): Unsubscribable & ProxyMarked
}

/** @internal */
export class ClientLanguageFeatures implements ClientLanguageFeaturesAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    constructor(
        private hoverRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideTextDocumentHoverSignature
        >,
        private definitionRegistry: TextDocumentLocationProviderRegistry,
        private referencesRegistry: TextDocumentLocationProviderRegistry<ReferenceParams>,
        private locationRegistry: TextDocumentLocationProviderIDRegistry,
        private completionItemsRegistry: FeatureProviderRegistry<
            TextDocumentRegistrationOptions,
            ProvideCompletionItemSignature
        >
    ) {}

    public $registerHoverProvider(
        documentSelector: DocumentSelector,
        providerFunction: Remote<
            ((params: TextDocumentPositionParams) => ProxySubscribable<Hover | null | undefined>) & ProxyMarked
        >
    ): Unsubscribable & ProxyMarked {
        return registerRemoteProvider(this.hoverRegistry, { documentSelector }, providerFunction)
    }

    public $registerDefinitionProvider(
        documentSelector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked {
        return registerRemoteProvider(this.definitionRegistry, { documentSelector }, providerFunction)
    }

    public $registerReferenceProvider(
        documentSelector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked {
        return registerRemoteProvider(this.referencesRegistry, { documentSelector }, providerFunction)
    }

    public $registerLocationProvider(
        id: string,
        documentSelector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked {
        return registerRemoteProvider(this.locationRegistry, { id, documentSelector }, providerFunction)
    }

    public $registerCompletionItemProvider(
        documentSelector: DocumentSelector,
        providerFunction: Remote<
            ((params: TextDocumentPositionParams) => ProxySubscribable<CompletionList | null | undefined>) & ProxyMarked
        >
    ): Unsubscribable & ProxyMarked {
        return registerRemoteProvider(this.completionItemsRegistry, { documentSelector }, providerFunction)
    }
}
