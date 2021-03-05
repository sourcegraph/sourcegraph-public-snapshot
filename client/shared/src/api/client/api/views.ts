import * as comlink from 'comlink'
import { isEqual, omit } from 'lodash'
import { combineLatest, ReplaySubject, Unsubscribable, ObservableInput, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { PanelView, View } from 'sourcegraph'
import { ContributableViewContainer } from '../../protocol'
import { TextDocumentLocationProviderIDRegistry } from '../services/location'
import { PanelViewWithComponent, PanelViewProviderRegistry } from '../services/panelViews'
import { Location } from '@sourcegraph/extension-api-types'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { ProxySubscribable } from '../../extension/api/common'
import { wrapRemoteObservable, ProxySubscription } from './common'
import { ViewService, ViewContexts } from '../services/viewService'
import { ExtensionHostAPI } from '../../extension/api/api'

/** @internal */
export interface PanelViewData extends Pick<PanelView, 'title' | 'content' | 'priority' | 'component'> {}

export interface PanelUpdater extends Unsubscribable, comlink.ProxyMarked {
    update(data: PanelViewData): void
}

/** @internal */
export interface ClientViewsAPI extends comlink.ProxyMarked {
    $registerDirectoryViewProvider(
        id: string,
        provider: comlink.Remote<
            ((context: ViewContexts[typeof ContributableViewContainer.Directory]) => ProxySubscribable<View | null>) &
                comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked

    $registerHomepageViewProvider(
        id: string,
        provider: comlink.Remote<
            ((context: ViewContexts[typeof ContributableViewContainer.Homepage]) => ProxySubscribable<View | null>) &
                comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked

    $registerInsightsPageViewProvider(
        id: string,
        provider: comlink.Remote<
            ((
                context: ViewContexts[typeof ContributableViewContainer.InsightsPage]
            ) => ProxySubscribable<View | null>) &
                comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked

    $registerGlobalPageViewProvider(
        id: string,
        provider: comlink.Remote<
            ((context: ViewContexts[typeof ContributableViewContainer.GlobalPage]) => ProxySubscribable<View | null>) &
                comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked
}

/** @internal */
export class ClientViews implements ClientViewsAPI {
    public readonly [comlink.proxyMarker] = true

    constructor(private viewService: ViewService) {}

    public $registerDirectoryViewProvider(
        id: string,
        provider: comlink.Remote<
            (
                context: ViewContexts[typeof ContributableViewContainer.Directory]
            ) => ProxySubscribable<View | null> & comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.viewService.register(id, ContributableViewContainer.Directory, context =>
                wrapRemoteObservable(provider(context), subscription)
            )
        )
        subscription.add(new ProxySubscription(provider))
        return comlink.proxy(subscription)
    }

    public $registerHomepageViewProvider(
        id: string,
        provider: comlink.Remote<
            (
                context: ViewContexts[typeof ContributableViewContainer.Homepage]
            ) => ProxySubscribable<View | null> & comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.viewService.register(id, ContributableViewContainer.Homepage, context =>
                wrapRemoteObservable(provider(context), subscription)
            )
        )
        subscription.add(new ProxySubscription(provider))
        return comlink.proxy(subscription)
    }

    public $registerInsightsPageViewProvider(
        id: string,
        provider: comlink.Remote<
            (
                context: ViewContexts[typeof ContributableViewContainer.InsightsPage]
            ) => ProxySubscribable<View | null> & comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.viewService.register(id, ContributableViewContainer.InsightsPage, context =>
                wrapRemoteObservable(provider(context), subscription)
            )
        )
        subscription.add(new ProxySubscription(provider))
        return comlink.proxy(subscription)
    }

    public $registerGlobalPageViewProvider(
        id: string,
        provider: comlink.Remote<
            (
                context: ViewContexts[typeof ContributableViewContainer.GlobalPage]
            ) => ProxySubscribable<View | null> & comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked {
        const subscription = new Subscription()
        subscription.add(
            this.viewService.register(id, ContributableViewContainer.GlobalPage, context =>
                wrapRemoteObservable(provider(context), subscription)
            )
        )
        subscription.add(new ProxySubscription(provider))
        return comlink.proxy(subscription)
    }
}
