import { Remote, ProxyMarked, proxy, proxyMarker } from 'comlink'
import { Hover, Location } from '@sourcegraph/extension-api-types'
import { CompletionList, DocumentSelector, Unsubscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { ReferenceParams, TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { ProvideCompletionItemSignature } from '../services/completion'
import { ProvideTextDocumentHoverSignature } from '../services/hover'
import { TextDocumentLocationProviderIDRegistry, TextDocumentLocationProviderRegistry } from '../services/location'
import { FeatureProviderRegistry } from '../services/registry'
import { wrapRemoteObservable, ProxySubscription } from './common'
import { Subscription } from 'rxjs'

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
        const subscription = new Subscription()
        subscription.add(
            this.hoverRegistry.registerProvider({ documentSelector }, params => {
                const remoteObservable = wrapRemoteObservable(providerFunction(params))
                subscription.add(remoteObservable.proxySubscription)
                return remoteObservable
            })
        )
        subscription.add(new ProxySubscription(providerFunction))
        return proxy(subscription)
    }

    public $registerDefinitionProvider(
        documentSelector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.definitionRegistry.registerProvider({ documentSelector }, params => {
                const remoteObservable = wrapRemoteObservable(providerFunction(params))
                subscription.add(remoteObservable.proxySubscription)
                return remoteObservable
            })
        )
        subscription.add(new ProxySubscription(providerFunction))
        return proxy(subscription)
    }

    public $registerReferenceProvider(
        documentSelector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.referencesRegistry.registerProvider({ documentSelector }, params => {
                const remoteObservable = wrapRemoteObservable(providerFunction(params))
                subscription.add(remoteObservable.proxySubscription)
                return remoteObservable
            })
        )
        subscription.add(new ProxySubscription(providerFunction))
        return proxy(subscription)
    }

    public $registerLocationProvider(
        id: string,
        documentSelector: DocumentSelector,
        providerFunction: Remote<((params: TextDocumentPositionParams) => ProxySubscribable<Location[]>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.locationRegistry.registerProvider({ id, documentSelector }, params => {
                const remoteObservable = wrapRemoteObservable(providerFunction(params))
                subscription.add(remoteObservable.proxySubscription)
                return remoteObservable
            })
        )
        subscription.add(new ProxySubscription(providerFunction))
        return proxy(subscription)
    }

    public $registerCompletionItemProvider(
        documentSelector: DocumentSelector,
        providerFunction: Remote<
            ((params: TextDocumentPositionParams) => ProxySubscribable<CompletionList | null | undefined>) & ProxyMarked
        >
    ): Unsubscribable & ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.completionItemsRegistry.registerProvider({ documentSelector }, params => {
                const remoteObservable = wrapRemoteObservable(providerFunction(params))
                subscription.add(remoteObservable.proxySubscription)
                return remoteObservable
            })
        )
        subscription.add(new ProxySubscription(providerFunction))
        return proxy(subscription)
    }
}
