import { ProxyResult, ProxyValue, proxy, proxyMarker } from '@sourcegraph/comlink'
import { LinkPreview, Unsubscribable } from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { LinkPreviewProviderRegistry } from '../services/linkPreview'
import { wrapRemoteObservable } from './common'

/** @internal */
export interface ClientContentAPI extends ProxyValue {
    $registerLinkPreviewProvider(
        urlMatchPattern: string,
        provider: ProxyResult<((url: string) => ProxySubscribable<LinkPreview | null | undefined>) & ProxyValue>
    ): Unsubscribable & ProxyValue
}

/** @internal */
export function createClientContent(registry: LinkPreviewProviderRegistry): ClientContentAPI & ProxyValue {
    return {
        [proxyMarker]: true,
        $registerLinkPreviewProvider: (urlMatchPattern, providerFunction) =>
            proxy(registry.registerProvider({ urlMatchPattern }, url => wrapRemoteObservable(providerFunction(url)))),
    }
}
