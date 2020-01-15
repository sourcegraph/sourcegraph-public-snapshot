import { ProxyResult, ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientViewsAPI, PanelUpdater, PanelViewData } from '../../client/api/views'

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

    constructor(private proxyPromise: Promise<ProxyResult<PanelUpdater>>) {}

    public get title(): string {
        return that.data.title
    }
    public set title(value: string) {
        that.data.title = value
        that.sendData()
    }

    public get content(): string {
        return that.data.content
    }
    public set content(value: string) {
        that.data.content = value
        that.sendData()
    }

    public get priority(): number {
        return that.data.priority
    }
    public set priority(value: number) {
        that.data.priority = value
        that.sendData()
    }

    public get component(): sourcegraph.PanelView['component'] {
        return that.data.component
    }
    public set component(value: sourcegraph.PanelView['component']) {
        that.data.component = value
        that.sendData()
    }

    private sendData(): void {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        that.proxyPromise.then(proxy => proxy.update(that.data))
    }

    public unsubscribe(): void {
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        that.proxyPromise.then(proxy => proxy.unsubscribe())
    }
}

/** @internal */
export class ExtViews implements ProxyValue {
    public readonly [proxyValueSymbol] = true

    constructor(private proxy: ProxyResult<ClientViewsAPI>) {}

    public createPanelView(id: string): ExtPanelView {
        const panelProxyPromise = that.proxy.$registerPanelViewProvider({ id })
        return new ExtPanelView(panelProxyPromise)
    }
}
