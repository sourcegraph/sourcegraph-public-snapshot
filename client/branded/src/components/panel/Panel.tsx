import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'
import { BehaviorSubject, from, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { Location } from '@sourcegraph/extension-api-types'
import { ActionsNavItems } from '@sourcegraph/shared/src/actions/ActionsNavItems'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { PanelViewData } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ContributableMenu, Contributions, Evaluated } from '@sourcegraph/shared/src/api/protocol'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { FetchFileParameters } from '@sourcegraph/shared/src/components/CodeExcerpt'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { combineLatestOrDefault } from '@sourcegraph/shared/src/util/rxjs/combineLatestOrDefault'
import { isDefined } from '@sourcegraph/shared/src/util/types'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import styles from './Panel.module.scss'
import { registerPanelToolbarContributions } from './views/contributions'
import { EmptyPanelView } from './views/EmptyPanelView'
import { ExtensionsLoadingPanelView } from './views/ExtensionsLoadingView'
import { PanelView } from './views/PanelView'

interface Props
    extends ExtensionsControllerProps,
        PlatformContextProps,
        SettingsCascadeProps,
        ActivationProps,
        TelemetryProps,
        ThemeProps {
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
interface PanelItem {
    id: string

    label: React.ReactFragment
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

    const [tabIndex, setTabIndex] = useState(0)
    const location = useLocation()
    const { hash, pathname, search } = location
    const history = useHistory()
    const handlePanelClose = useCallback(() => history.replace(pathname), [history, pathname])
    const [currentTabLabel, currentTabID] = hash.split('=')

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
                        panelViews.map((panelView: PanelViewWithComponent) => {
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

    const panelViews = useMemo(() => [...(builtinPanels || []), ...(extensionPanels || [])], [
        builtinPanels,
        extensionPanels,
    ])

    const items = useMemo(
        () =>
            panelViews
                ? panelViews
                      .map(
                          (panelView): PanelItem => ({
                              label: panelView.title,
                              id: panelView.id,
                              priority: panelView.priority,
                              element: <PanelView {...props} panelView={panelView} location={location} />,
                              hasLocations: !!panelView.locationProvider,
                          })
                      )
                      .sort((a, b) => b.priority - a.priority)
                : [],
        [location, panelViews, props]
    )

    useEffect(() => {
        const subscription = registerPanelToolbarContributions(props.extensionsController.extHostAPI)
        return () => subscription.unsubscribe()
    }, [props.extensionsController])

    const handleActiveTab = useCallback(
        (index: number): void => {
            history.replace(`${pathname}${search}${currentTabLabel}=${items[index].id}`)
        },
        [currentTabLabel, history, items, pathname, search]
    )

    useEffect(() => {
        setTabIndex(items.findIndex(({ id }) => id === currentTabID))
    }, [items, hash, currentTabID])

    if (!areExtensionsReady) {
        return <ExtensionsLoadingPanelView className={styles.panel} />
    }

    if (!items) {
        return <EmptyPanelView className={styles.panel} />
    }

    const activeTab: PanelItem | undefined = items[tabIndex]

    return (
        <Tabs className={styles.panel} index={tabIndex} onChange={handleActiveTab}>
            <div className={classNames('tablist-wrapper d-flex justify-content-between sticky-top', styles.header)}>
                <TabList>
                    <div className="d-flex w-100">
                        {items.map(({ label, id }) => (
                            <Tab key={id}>
                                <span className="tablist-wrapper--tab-label">{label}</span>
                            </Tab>
                        ))}
                    </div>
                </TabList>
                <div className="align-items-center d-flex">
                    <small>
                        {activeTab && (
                            <ActionsNavItems
                                {...props}
                                // TODO remove references to Bootstrap from shared, get class name from prop
                                // This is okay for now because the Panel is currently only used in the webapp
                                listClass="d-flex justify-content-end list-unstyled m-0 align-items-center"
                                listItemClass="px-2 mx-2"
                                actionItemClass="font-weight-medium"
                                actionItemIconClass="icon-inline"
                                menu={ContributableMenu.PanelToolbar}
                                scope={{
                                    type: 'panelView',
                                    id: activeTab.id,
                                    hasLocations: Boolean(activeTab.hasLocations),
                                }}
                                wrapInList={true}
                                location={location}
                                transformContributions={transformPanelContributions}
                            />
                        )}
                    </small>
                    <button
                        type="button"
                        onClick={handlePanelClose}
                        className={classNames('btn btn-icon ml-2', styles.dismissButton)}
                        title="Close panel"
                        data-tooltip="Close panel"
                        data-placement="left"
                    >
                        <CloseIcon className="icon-inline" />
                    </button>
                </div>
            </div>
            <TabPanels className={styles.tabs}>
                {activeTab ? (
                    items.map(({ id, element }) => (
                        <TabPanel key={id} className={styles.tabsContent} data-testid="panel-tabs-content">
                            {id === activeTab.id ? element : null}
                        </TabPanel>
                    ))
                ) : (
                    <EmptyPanelView />
                )}
            </TabPanels>
        </Tabs>
    )
})

/** A wrapper around Panel that makes it resizable. */
export const ResizablePanel: React.FunctionComponent<Props> = props => (
    <Resizable
        className={styles.resizablePanel}
        handlePosition="top"
        defaultSize={350}
        storageKey="panel-size"
        element={<Panel {...props} />}
    />
)

/**
 * Temporary solution to code intel extensions all contributing the same panel actions.
 */
function transformPanelContributions(contributions: Evaluated<Contributions>): Evaluated<Contributions> {
    try {
        const panelMenuItems = contributions.menus?.['panel/toolbar']
        if (!panelMenuItems || panelMenuItems.length === 0) {
            return contributions
        }
        // This won't dedup e.g. [{action: 'a', when: 'b'}, {when: 'b', action: 'a'}], but should
        // work for the case this is hackily trying to prevent: multiple extensions generated with the
        // same manifest.
        const strings = new Set(panelMenuItems.map(menuItem => JSON.stringify(menuItem)))
        const uniquePanelMenuItems = [...strings].map(string => JSON.parse(string))

        return {
            ...contributions,
            menus: {
                ...contributions.menus,
                'panel/toolbar': uniquePanelMenuItems,
            },
        }
    } catch {
        return contributions
    }
}
