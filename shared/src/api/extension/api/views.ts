import { Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientViewsAPI, PanelViewData } from '../../client/api/views'
import { ProviderMap } from './common'

/**
 * @internal
 */
class ExtPanelView implements sourcegraph.PanelView {
    private data: PanelViewData = {
        title: '',
        content: '',
        priority: 0,
    }

    constructor(private proxy: ClientViewsAPI, private id: number, private subscription: Unsubscribable) {}

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

    private sendData(): void {
        this.proxy.$acceptPanelViewUpdate(this.id, this.data)
    }

    public unsubscribe(): void {
        return this.subscription.unsubscribe()
    }
}

/** @internal */
export class ExtViews implements Unsubscribable {
    private registrations = new ProviderMap<{}>(id => this.proxy.$unregister(id))

    constructor(private proxy: ClientViewsAPI) {}

    public createPanelView(id: string): ExtPanelView {
        const { id: regID, subscription } = this.registrations.add({})
        this.proxy.$registerPanelViewProvider(regID, { id })
        return new ExtPanelView(this.proxy, regID, subscription)
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}
