import { combineLatest, ReplaySubject, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { handleRequests } from '../../common/proxy'
import { ContributableViewContainer } from '../../protocol'
import { Connection } from '../../protocol/jsonrpc2/connection'
import * as plain from '../../protocol/plainTypes'
import { ViewProviderRegistry } from '../providers/view'
import { SubscriptionMap } from './common'

/** @internal */
export interface ClientViewsAPI {
    $unregister(id: number): void
    $registerPanelViewProvider(id: number, provider: { id: string }): void
    $acceptPanelViewUpdate(id: number, params: Partial<plain.PanelView>): void
}

interface PanelViewSubjects {
    title: Subject<string>
    content: Subject<string>
}

/** @internal */
export class ClientViews implements ClientViewsAPI {
    private subscriptions = new Subscription()
    private panelViews = new Map<number, Record<keyof plain.PanelView, Subject<string>>>()
    private registrations = new SubscriptionMap()

    constructor(connection: Connection, private viewRegistry: ViewProviderRegistry) {
        this.subscriptions.add(this.registrations)

        handleRequests(connection, 'views', this)
    }

    public $unregister(id: number): void {
        this.registrations.remove(id)
    }

    public $registerPanelViewProvider(id: number, provider: { id: string }): void {
        const panelView: PanelViewSubjects = {
            title: new ReplaySubject<string>(1),
            content: new ReplaySubject<string>(1),
        }
        this.panelViews.set(id, panelView)
        const registryUnsubscribable = this.viewRegistry.registerProvider(
            { ...provider, container: ContributableViewContainer.Panel },
            combineLatest(panelView.title, panelView.content).pipe(map(([title, content]) => ({ title, content })))
        )
        this.registrations.add(id, {
            unsubscribe: () => {
                registryUnsubscribable.unsubscribe()
                this.panelViews.delete(id)
            },
        })
    }

    public $acceptPanelViewUpdate(id: number, params: { title?: string; content?: string }): void {
        const panelView = this.panelViews.get(id)
        if (panelView === undefined) {
            throw new Error(`no panel view with ID ${id}`)
        }
        if (params.title !== undefined) {
            panelView.title.next(params.title)
        }
        if (params.content !== undefined) {
            panelView.content.next(params.content)
        }
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
