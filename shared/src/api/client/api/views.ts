import * as comlink from '@sourcegraph/comlink'
import { isEqual, omit } from 'lodash'
import { combineLatest, from, ReplaySubject, Unsubscribable, ObservableInput } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { PanelView, View } from 'sourcegraph'
import { ContributableViewContainer } from '../../protocol'
import { EditorService, getActiveCodeEditorPosition } from '../services/editorService'
import { TextDocumentLocationProviderIDRegistry } from '../services/location'
import { PanelViewWithComponent, PanelViewProviderRegistry } from '../services/panelViews'
import { Location } from '@sourcegraph/extension-api-types'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { ProxySubscribable } from '../../extension/api/common'
import { wrapRemoteObservable } from './common'
import { ViewService } from '../services/viewService'

/** @internal */
export interface PanelViewData extends Pick<PanelView, 'title' | 'content' | 'priority' | 'component'> {}

export interface PanelUpdater extends Unsubscribable, comlink.ProxyMarked {
    update(data: PanelViewData): void
}

/** @internal */
export interface ClientViewsAPI extends comlink.ProxyMarked {
    $registerPanelViewProvider(provider: { id: string }): PanelUpdater

    $registerViewProvider(
        id: string,
        providerFunction: comlink.Remote<
            ((params: { [key: string]: string }) => ProxySubscribable<View>) & comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked
}

/** @internal */
export class ClientViews implements ClientViewsAPI {
    public readonly [comlink.proxyMarker] = true

    constructor(
        private panelViewRegistry: PanelViewProviderRegistry,
        private textDocumentLocations: TextDocumentLocationProviderIDRegistry,
        private editorService: EditorService,
        private viewService: ViewService
    ) {}

    public $registerPanelViewProvider(provider: { id: string }): PanelUpdater {
        // TODO(sqs): This will probably hang forever if an extension neglects to set any of the fields on a
        // PanelView because this subject will never emit.
        const panelView = new ReplaySubject<PanelViewData>(1)
        const registryUnsubscribable = this.panelViewRegistry.registerProvider(
            { ...provider, container: ContributableViewContainer.Panel },
            combineLatest([
                panelView.pipe(
                    map(data => omit(data, 'component')),
                    distinctUntilChanged((x, y) => isEqual(x, y))
                ),
                panelView.pipe(
                    map(({ component }) => component),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    map(component => {
                        if (!component) {
                            return undefined
                        }

                        return from(this.editorService.activeEditorUpdates).pipe(
                            map(getActiveCodeEditorPosition),
                            switchMap(
                                (params): ObservableInput<MaybeLoadingResult<Location[]>> => {
                                    if (!params) {
                                        return [{ isLoading: false, result: [] }]
                                    }
                                    return this.textDocumentLocations.getLocations(component.locationProvider, params)
                                }
                            )
                        )
                    })
                ),
            ]).pipe(
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
        return comlink.proxy({
            update: (data: PanelViewData) => {
                panelView.next(data)
            },
            unsubscribe: () => {
                registryUnsubscribable.unsubscribe()
            },
        })
    }

    public $registerViewProvider(
        id: string,
        providerFunction: comlink.Remote<
            ((params: Record<string, string>) => ProxySubscribable<View>) & comlink.ProxyMarked
        >
    ): Unsubscribable & comlink.ProxyMarked {
        return comlink.proxy(this.viewService.register(id, params => wrapRemoteObservable(providerFunction(params))))
    }
}
