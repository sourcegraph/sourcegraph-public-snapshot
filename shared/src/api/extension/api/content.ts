import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import { Unsubscribable } from 'rxjs'
import { LinkPreviewProvider } from 'sourcegraph'
import { ClientContentAPI } from '../../client/api/content'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'

/** @internal */
export class ExtContent {
    constructor(private proxy: ProxyResult<ClientContentAPI>) {}

    public registerLinkPreviewProvider(urlMatchPattern: string, provider: LinkPreviewProvider): Unsubscribable {
        const providerFunction: ProxyInput<Parameters<
            ClientContentAPI['$registerLinkPreviewProvider']
        >[1]> = proxyValue((url: string) =>
            toProxyableSubscribable(provider.provideLinkPreview(new URL(url)), preview => preview)
        )
        return syncSubscription(this.proxy.$registerLinkPreviewProvider(urlMatchPattern, providerFunction))
    }
}
