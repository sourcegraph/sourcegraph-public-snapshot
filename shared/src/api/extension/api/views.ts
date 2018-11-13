import { Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ClientViewsAPI } from '../../client/api/views'
import { ProviderMap } from './common'

/**
 * @internal
 */
class ExtPanelView implements sourcegraph.PanelView {
    private _title = ''
    private _content = ''

    constructor(private proxy: ClientViewsAPI, private id: number, private subscription: Unsubscribable) {}

    public get title(): string {
        return this._title
    }
    public set title(value: string) {
        this._title = value
        this.proxy.$acceptPanelViewUpdate(this.id, { title: value })
    }

    public get content(): string {
        return this._content
    }
    public set content(value: string) {
        this._content = value
        this.proxy.$acceptPanelViewUpdate(this.id, { content: value })
    }

    public unsubscribe(): void {
        return this.subscription.unsubscribe()
    }
}

/** @internal */
export interface ExtViewsAPI {}

/** @internal */
export class ExtViews implements ExtViewsAPI {
    private registrations = new ProviderMap<{}>(id => this.proxy.$unregister(id))

    constructor(private proxy: ClientViewsAPI) {}

    public createPanelView(id: string): ExtPanelView {
        const { id: regID, subscription } = this.registrations.add({})
        this.proxy.$registerPanelViewProvider(regID, { id })
        return new ExtPanelView(this.proxy, regID, subscription)
    }
}
