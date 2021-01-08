import * as H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useMemo } from 'react'
import { BehaviorSubject, from, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { ContributableMenu } from '../../../../shared/src/api/protocol/contribution'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { ActionsNavItems } from '../../../../shared/src/actions/ActionsNavItems'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { FetchFileParameters } from '../../../../shared/src/components/CodeExcerpt'
import { Resizable } from '../../../../shared/src/components/Resizable'
import { Spacer, Tab, TabsWithURLViewStatePersistence } from '../Tabs'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { EmptyPanelView } from './views/EmptyPanelView'
import { PanelView } from './views/PanelView'
import { ThemeProps } from '../../../../shared/src/theme'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { combineLatestOrDefault } from '../../../../shared/src/util/rxjs/combineLatestOrDefault'
import { Location } from '@sourcegraph/extension-api-types'
import { isDefined } from '../../../../shared/src/util/types'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { wrapRemoteObservable } from '../../../../shared/src/api/client/api/common'
import { PanelViewData } from '../../../../shared/src/api/extension/flatExtensionApi'
import { ExtensionsLoadingPanelView } from './views/ExtensionsLoadingView'
import { haveInitialExtensionsLoaded } from '../../../../shared/src/api/features'

interface Props
    extends ExtensionsControllerProps,
        PlatformContextProps,
        SettingsCascadeProps,
        ActivationProps,
        TelemetryProps,
        ThemeProps,
        VersionContextProps {
    location: H.Location
    history: H.History
    repoName?: string
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export interface PanelViewWithComponent extends PanelViewData {
    /**
     * The location provider whose results to render in the panel view.
     */
    locationProvider?: Observable<MaybeLoadingResult<Location[]>>

    /**
     * The React element to render in the panel view.
     */
    reactElement?: React.ReactFragment
}

/**
 * A tab and corresponding content to display in the panel.
 */
interface PanelItem extends Tab<string> {
    /**
     * Controls the relative order of panel items. The items are laid out from highest priority (at the beginning)
     * to lowest priority (at the end). The default is 0.
     */
    priority: number

    /** The content element to display when the tab is active. */
    element: JSX.Element

    /**
     * Whether this panel contains a list of locations (from a location provider). This value is
     * exposed to contributions as `panel.activeView.hasLocations`. It is true if there is a
     * location provider (even if the result set is empty).
     */
    hasLocations?: boolean
}

export type BuiltinPanelView = Omit<PanelViewWithComponent, 'component' | 'id'>

const builtinPanelViewProviders = new BehaviorSubject<
    Map<string, { id: string; provider: Observable<BuiltinPanelView | null> }>
>(new Map())

/**
 * React hook to add panel views from other components (panel views are typically
 * contributed by Sourcegraph extensions)
 */
export function useBuiltinPanelViews(
    builtinPanels: { id: string; provider: Observable<BuiltinPanelView | null> }[]
): void {
    useEffect(() => {
        for (const builtinPanel of builtinPanels) {
            builtinPanelViewProviders.value.set(builtinPanel.id, builtinPanel)
        }
        builtinPanelViewProviders.next(new Map([...builtinPanelViewProviders.value]))

        return () => {
            for (const builtinPanel of builtinPanels) {
                builtinPanelViewProviders.value.delete(builtinPanel.id)
            }
            builtinPanelViewProviders.next(new Map([...builtinPanelViewProviders.value]))
        }
    }, [builtinPanels])
}

/**
 * The panel, which is a tabbed component with contextual information. Components rendering the panel should
 * generally use ResizablePanel, not Panel.
 *
 * Other components can contribute panel items to the panel with the `useBuildinPanelViews` hook.
 */
export const Panel = React.memo<Props>(props => {
    // Ensures that we don't show a misleading empty state when extensions haven't loaded yet.
    const areExtensionsReady = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(props.extensionsController.extHostAPI), [props.extensionsController])
    )

    const builtinPanels: PanelViewWithComponent[] | undefined = useObservable(
        useMemo(
            () =>
                builtinPanelViewProviders.pipe(
                    switchMap(providers =>
                        combineLatestOrDefault(
                            [...providers].map(([id, { provider }]) =>
                                provider.pipe(map(view => (view ? { ...view, id, component: null } : null)))
                            )
                        )
                    ),
                    map(views => views.filter(isDefined))
                ),
            []
        )
    )

    const extensionPanels: PanelViewWithComponent[] | undefined = useObservable(
        useMemo(
            () =>
                from(props.extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI =>
                        wrapRemoteObservable(extensionHostAPI.getPanelViews()).pipe(
                            map(panelViews => ({ panelViews, extensionHostAPI }))
                        )
                    ),
                    map(({ panelViews, extensionHostAPI }) =>
                        panelViews.map(panelView => {
                            const locationProviderID = panelView.component?.locationProvider
                            if (locationProviderID) {
                                const panelViewWithProvider: PanelViewWithComponent = {
                                    ...panelView,
                                    locationProvider: wrapRemoteObservable(
                                        extensionHostAPI.getActiveCodeEditorPosition()
                                    ).pipe(
                                        switchMap(parameters => {
                                            if (!parameters) {
                                                return [{ isLoading: false, result: [] }]
                                            }

                                            return wrapRemoteObservable(
                                                extensionHostAPI.getLocations(locationProviderID, parameters)
                                            )
                                        })
                                    ),
                                }
                                return panelViewWithProvider
                            }

                            return panelView
                        })
                    )
                ),
            [props.extensionsController]
        )
    )

    const onDismiss = useCallback(
        () => props.history.push(TabsWithURLViewStatePersistence.urlForTabID(props.location, null)),
        [props.history, props.location]
    )

    const panelViews = [...(builtinPanels || []), ...(extensionPanels || [])]

    const items = panelViews
        ? panelViews
              .map(
                  (panelView): PanelItem => ({
                      label: panelView.title,
                      id: panelView.id,
                      priority: panelView.priority,
                      element: <PanelView {...props} panelView={panelView} />,
                      hasLocations: !!panelView.locationProvider,
                  })
              )
              .sort(byPriority)
        : []
    const hasTabs = items.length > 0
    const activePanelViewID = TabsWithURLViewStatePersistence.readFromURL(props.location, items)
    const activePanelView = items.find(item => item.id === activePanelViewID)

    return (
        <div className="panel">
            {!areExtensionsReady ? (
                <ExtensionsLoadingPanelView />
            ) : hasTabs ? (
                <TabsWithURLViewStatePersistence
                    tabs={items}
                    tabBarEndFragment={
                        <>
                            <Spacer />
                            <ActionsNavItems
                                {...props}
                                // TODO remove references to Bootstrap from shared, get class name from prop
                                // This is okay for now because the Panel is currently only used in the webapp
                                listClass="nav panel__actions"
                                actionItemClass="nav-link mw-100 panel__action"
                                actionItemIconClass="icon-inline"
                                menu={ContributableMenu.PanelToolbar}
                                scope={
                                    activePanelView !== undefined
                                        ? {
                                              type: 'panelView',
                                              id: activePanelView.id,
                                              hasLocations: Boolean(activePanelView.hasLocations),
                                          }
                                        : undefined
                                }
                                wrapInList={true}
                            />
                            <button
                                type="button"
                                onClick={onDismiss}
                                className="btn btn-icon tab-bar__end-fragment-other-element panel__dismiss"
                                data-tooltip="Close"
                            >
                                <CloseIcon className="icon-inline" />
                            </button>
                        </>
                    }
                    className="panel__tabs"
                    tabBarClassName="panel__tab-bar"
                    tabClassName="tab-bar__tab--h5like"
                    location={props.location}
                >
                    {items?.map(({ id, element }) => React.cloneElement(element, { key: id }))}
                </TabsWithURLViewStatePersistence>
            ) : (
                <EmptyPanelView />
            )}
        </div>
    )
})

function byPriority(a: { priority: number }, b: { priority: number }): number {
    return b.priority - a.priority
}

/** A wrapper around Panel that makes it resizable. */
export const ResizablePanel: React.FunctionComponent<Props> = props => (
    <Resizable
        className="resizable-panel"
        handlePosition="top"
        defaultSize={350}
        storageKey="panel-size"
        element={<Panel {...props} />}
    />
)
