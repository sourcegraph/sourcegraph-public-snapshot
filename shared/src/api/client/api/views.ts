import { ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { ReplaySubject, Unsubscribable } from 'rxjs'
import { map } from 'rxjs/operators'
import { PanelView } from 'sourcegraph'
import { ContributableViewContainer, TextDocumentPositionParams } from '../../protocol'
import { TextDocumentLocationProviderIDRegistry } from '../services/location'
import { PanelViewWithComponent, ViewProviderRegistry } from '../services/view'

/** @internal */
export interface PanelViewData extends Pick<PanelView, 'title' | 'content' | 'priority' | 'component'> {}

export interface PanelUpdater extends Unsubscribable, ProxyValue {
    update(data: PanelViewData): void
}

/** @internal */
export interface ClientViewsAPI extends ProxyValue {
    $registerPanelViewProvider(provider: { id: string }): PanelUpdater
}

/** @internal */
export class ClientViews implements ClientViewsAPI {
    public readonly [proxyValueSymbol] = true

    constructor(
        private viewRegistry: ViewProviderRegistry,
        private textDocumentLocations: TextDocumentLocationProviderIDRegistry
    ) {}

    public $registerPanelViewProvider(provider: { id: string }): PanelUpdater {
        // TODO(sqs): This will probably hang forever if an extension neglects to set any of the fields on a
        // PanelView because this subject will never emit.
        const panelView = new ReplaySubject<PanelViewData>(1)
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
        return proxyValue({
            update: (data: PanelViewData) => {
                panelView.next(data)
            },
            unsubscribe: () => {
                registryUnsubscribable.unsubscribe()
            },
        })
    }
}
