import { ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { isEqual } from 'lodash'
import { combineLatest, from, of, ReplaySubject, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, map, switchMap } from 'rxjs/operators'
import { PanelView } from 'sourcegraph'
import { ContributableViewContainer, TextDocumentPositionParams } from '../../protocol'
import { modelToTextDocumentPositionParams } from '../model'
import { TextDocumentLocationProviderIDRegistry } from '../services/location'
import { ModelService } from '../services/modelService'
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
        private textDocumentLocations: TextDocumentLocationProviderIDRegistry,
        private modelService: ModelService
    ) {}

    public $registerPanelViewProvider(provider: { id: string }): PanelUpdater {
        // TODO(sqs): This will probably hang forever if an extension neglects to set any of the fields on a
        // PanelView because this subject will never emit.
        const panelView = new ReplaySubject<PanelViewData>(1)
        const registryUnsubscribable = this.viewRegistry.registerProvider(
            { ...provider, container: ContributableViewContainer.Panel },
            combineLatest(
                panelView.pipe(
                    map(({ title, content, priority }) => ({ title, content, priority })),
                    distinctUntilChanged((x, y) => isEqual(x, y))
                ),
                panelView.pipe(
                    map(({ component }) => component),
                    filter((component): component is { locationProvider: string } => Boolean(component)),
                    map(({ locationProvider }) => locationProvider),
                    distinctUntilChanged(),
                    map(locationProvider =>
                        from(this.modelService.model).pipe(
                            switchMap(model => {
                                const params: TextDocumentPositionParams | null = modelToTextDocumentPositionParams(
                                    model
                                )
                                if (!params) {
                                    return of(of(null))
                                }
                                return this.textDocumentLocations.getLocations(locationProvider, params)
                            })
                        )
                    )
                )
            ).pipe(
                map(([{ title, content, priority }, locationProvider]) => {
                    const panelView: PanelViewWithComponent = {
                        title,
                        content,
                        priority,
                        locationProvider,
                    }
                    return panelView
                })
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
