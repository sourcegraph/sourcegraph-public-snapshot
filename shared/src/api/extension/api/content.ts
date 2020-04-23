import { Local, Remote, proxy } from '@sourcegraph/comlink'
import { Unsubscribable } from 'rxjs'
import { LinkPreviewProvider } from 'sourcegraph'
import { ClientContentAPI } from '../../client/api/content'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'

/** @internal */
export class ExtContent {
    constructor(private proxy: Remote<ClientContentAPI>) {}

    public registerLinkPreviewProvider(urlMatchPattern: string, provider: LinkPreviewProvider): Unsubscribable {
        const providerFunction: Local<
            Parameters<ClientContentAPI['$registerLinkPreviewProvider']>[1]
        > = proxy((url: string) =>
            toProxyableSubscribable(provider.provideLinkPreview(new URL(url)), preview => preview)
        )
        return syncSubscription(this.proxy.$registerLinkPreviewProvider(urlMatchPattern, providerFunction))
    }
}
