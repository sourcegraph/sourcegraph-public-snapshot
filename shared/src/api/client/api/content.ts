import { Remote, ProxyMarked, proxy, proxyMarker, releaseProxy } from 'comlink'
import { LinkPreview, Unsubscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { LinkPreviewProviderRegistry } from '../services/linkPreview'
import { wrapRemoteObservable } from './common'
import { Subscription } from 'rxjs'

/** @internal */
export interface ClientContentAPI extends ProxyMarked {
    $registerLinkPreviewProvider(
        urlMatchPattern: string,
        provider: Remote<((url: string) => ProxySubscribable<LinkPreview | null | undefined>) & ProxyMarked>
    ): Unsubscribable & ProxyMarked
}

/** @internal */
export function createClientContent(registry: LinkPreviewProviderRegistry): ClientContentAPI & ProxyMarked {
    return {
        [proxyMarker]: true,
        $registerLinkPreviewProvider: (urlMatchPattern, providerFunction) => {
            const subscription = new Subscription()
            subscription.add(
                registry.registerProvider({ urlMatchPattern }, url => {
                    const remoteObservable = wrapRemoteObservable(providerFunction(url))
                    subscription.add(() => remoteObservable[releaseProxy]())
                    return remoteObservable
                })
            )
            subscription.add(() => providerFunction[releaseProxy]())
            return proxy(subscription)
        },
    }
}
