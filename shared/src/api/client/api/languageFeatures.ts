import { Remote, ProxyMarked, proxyMarker } from 'comlink'
import { Location } from '@sourcegraph/extension-api-types'
import { DocumentSelector, Unsubscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { ReferenceParams, TextDocumentPositionParams } from '../../protocol'
import { TextDocumentLocationProviderIDRegistry, TextDocumentLocationProviderRegistry } from '../services/location'
import { registerRemoteProvider } from './common'

/** @internal */
export interface ClientLanguageFeaturesAPI extends ProxyMarked {
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
}

/** @internal */
export class ClientLanguageFeatures implements ClientLanguageFeaturesAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    constructor(
        private definitionRegistry: TextDocumentLocationProviderRegistry,
        private referencesRegistry: TextDocumentLocationProviderRegistry<ReferenceParams>,
        private locationRegistry: TextDocumentLocationProviderIDRegistry
    ) {}

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
}
