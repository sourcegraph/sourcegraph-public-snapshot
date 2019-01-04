import { ReplaySubject, Subject, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { PanelView } from 'sourcegraph'
import { handleRequests } from '../../common/proxy'
import { ContributableViewContainer, TextDocumentPositionParams } from '../../protocol'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { TextDocumentLocationProviderIDRegistry } from '../services/location'
import { PanelViewWithComponent, ViewProviderRegistry } from '../services/view'
import { SubscriptionMap } from './common'

/** @internal */
export interface PanelViewData extends Pick<PanelView, 'title' | 'content' | 'priority' | 'component'> {}

/** @internal */
export interface ClientViewsAPI {
    $unregister(id: number): void
    $registerPanelViewProvider(id: number, provider: { id: string }): void
    $acceptPanelViewUpdate(id: number, params: Partial<PanelViewData>): void
}

/** @internal */
export class ClientViews implements ClientViewsAPI {
    private subscriptions = new Subscription()
    private panelViews = new Map<number, Subject<PanelViewData>>()
    private registrations = new SubscriptionMap()

    constructor(
        connection: Connection,
        private viewRegistry: ViewProviderRegistry,
        private textDocumentLocations: TextDocumentLocationProviderIDRegistry
    ) {
        this.subscriptions.add(this.registrations)

        handleRequests(connection, 'views', this)
    }

    public $unregister(id: number): void {
        this.registrations.remove(id)
    }

    public $registerPanelViewProvider(id: number, provider: { id: string }): void {
        // TODO(sqs): This will probably hang forever if an extension neglects to set any of the fields on a
        // PanelView because this subject will never emit.
        const panelView = new ReplaySubject<PanelViewData>(1)
        this.panelViews.set(id, panelView)
        const registryUnsubscribable = this.viewRegistry.registerProvider(
            { ...provider, container: ContributableViewContainer.Panel },
            panelView.pipe(
                map(
                    ({ title, content, priority, component }) =>
                        ({
                            title,
                            content,
                            priority,
                            locationProvider: component
                                ? (params: TextDocumentPositionParams) =>
                                      this.textDocumentLocations.getLocations(component.locationProvider, params)
                                : undefined,
                        } as PanelViewWithComponent)
                )
            )
        )
        this.registrations.add(id, {
            unsubscribe: () => {
                registryUnsubscribable.unsubscribe()
                this.panelViews.delete(id)
            },
        })
    }

    public $acceptPanelViewUpdate(id: number, data: PanelViewData): void {
        const panelView = this.panelViews.get(id)
        if (panelView === undefined) {
            throw new Error(`no panel view with ID ${id}`)
        }
        panelView.next(data)
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
