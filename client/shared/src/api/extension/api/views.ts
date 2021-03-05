import * as comlink from 'comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientViewsAPI } from '../../client/api/views'
import { syncSubscription } from '../../util'
import { Unsubscribable } from 'rxjs'
import { toProxyableSubscribable } from './common'
import { ContributableViewContainer } from '../../protocol'
import { ViewContexts } from '../../client/services/viewService'

/** @internal */
export class ExtensionViewsApi implements comlink.ProxyMarked {
    public readonly [comlink.proxyMarker] = true

    constructor(private proxy: comlink.Remote<ClientViewsAPI>) {}

    public registerViewProvider(id: string, provider: sourcegraph.ViewProvider): Unsubscribable {
        switch (provider.where) {
            case ContributableViewContainer.Directory: {
                return syncSubscription(
                    this.proxy.$registerDirectoryViewProvider(
                        id,
                        comlink.proxy((context: ViewContexts[typeof ContributableViewContainer.Directory]) =>
                            toProxyableSubscribable(
                                provider.provideView({
                                    viewer: {
                                        ...context.viewer,
                                        directory: {
                                            ...context.viewer.directory,
                                            uri: new URL(context.viewer.directory.uri),
                                        },
                                    },
                                    workspace: {
                                        uri: new URL(context.workspace.uri),
                                    },
                                }),
                                result => result || null
                            )
                        )
                    )
                )
            }
            case ContributableViewContainer.Homepage: {
                return syncSubscription(
                    this.proxy.$registerHomepageViewProvider(
                        id,
                        comlink.proxy((context: ViewContexts[typeof ContributableViewContainer.Homepage]) =>
                            toProxyableSubscribable(provider.provideView(context), result => result || null)
                        )
                    )
                )
            }
            case ContributableViewContainer.InsightsPage: {
                return syncSubscription(
                    this.proxy.$registerInsightsPageViewProvider(
                        id,
                        comlink.proxy((context: ViewContexts[typeof ContributableViewContainer.InsightsPage]) =>
                            toProxyableSubscribable(provider.provideView(context), result => result || null)
                        )
                    )
                )
            }
            case ContributableViewContainer.GlobalPage: {
                return syncSubscription(
                    this.proxy.$registerGlobalPageViewProvider(
                        id,
                        comlink.proxy((context: ViewContexts[typeof ContributableViewContainer.GlobalPage]) =>
                            toProxyableSubscribable(provider.provideView(context), result => result || null)
                        )
                    )
                )
            }
        }
    }
}
