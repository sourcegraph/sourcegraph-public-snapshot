import * as comlink from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientViewsAPI, PanelUpdater, PanelViewData } from '../../client/api/views'
import { syncSubscription } from '../../util'
import { Unsubscribable } from 'rxjs'
import { toProxyableSubscribable, ProxySubscribable } from './common'
import { ContributableViewContainer } from '../../protocol'

/**
 * @internal
 */
class ExtPanelView implements sourcegraph.PanelView {
    private data: PanelViewData = {
        title: '',
        content: '',
        priority: 0,
        component: null,
    }

    constructor(private proxyPromise: Promise<comlink.Remote<PanelUpdater>>) {}

    public get title(): string {
        return this.data.title
    }
    public set title(value: string) {
        this.data.title = value
        this.sendData()
    }

    public get content(): string {
        return this.data.content
    }
    public set content(value: string) {
        this.data.content = value
        this.sendData()
    }

    public get priority(): number {
        return this.data.priority
    }
    public set priority(value: number) {
        this.data.priority = value
        this.sendData()
    }

    public get component(): sourcegraph.PanelView['component'] {
        return this.data.component
    }
    public set component(value: sourcegraph.PanelView['component']) {
        this.data.component = value
        this.sendData()
    }

    private sendData(): void {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        this.proxyPromise.then(proxy => proxy.update(this.data))
    }

    public unsubscribe(): void {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        this.proxyPromise.then(proxy => proxy.unsubscribe())
    }
}

/** @internal */
export class ExtViews implements comlink.ProxyMarked {
    public readonly [comlink.proxyMarker] = true

    constructor(private proxy: comlink.Remote<ClientViewsAPI>) {}

    public createPanelView(id: string): ExtPanelView {
        const panelProxyPromise = this.proxy.$registerPanelViewProvider({ id })
        return new ExtPanelView(panelProxyPromise)
    }

    public registerViewProvider(id: string, provider: sourcegraph.ViewProvider): Unsubscribable {
        const providerFunction: comlink.Local<comlink.Remote<
            ((context: any) => ProxySubscribable<sourcegraph.View | null>) & comlink.ProxyMarked
        >> = comlink.proxy((context: any) => {
            const rehydrated =
                provider.where === 'directory'
                    ? { workspace: ((): sourcegraph.WorkspaceRoot => ({ uri: new URL(context.workspace.uri) }))() }
                    : context
            return toProxyableSubscribable(provider.provideView(rehydrated), result => result || null)
        })
        return syncSubscription(
            this.proxy.$registerViewProvider(id, provider.where as ContributableViewContainer, providerFunction)
        )
    }
}
