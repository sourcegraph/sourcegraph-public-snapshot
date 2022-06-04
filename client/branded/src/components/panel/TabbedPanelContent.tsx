import React, { useCallback, useEffect, useMemo, useState } from 'react'

import classNames from 'classnames'
import { Remote } from 'comlink'
import CloseIcon from 'mdi-react/CloseIcon'
import { useHistory, useLocation } from 'react-router'
import { BehaviorSubject, from, Observable, combineLatest } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { ContributableMenu, Contributions, Evaluated } from '@sourcegraph/client-api'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { isDefined, combineLatestOrDefault, isErrorLike } from '@sourcegraph/common'
import { Location } from '@sourcegraph/extension-api-types'
import { FetchFileParameters } from '@sourcegraph/search-ui'
import { ActionsNavItems } from '@sourcegraph/shared/src/actions/ActionsNavItems'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { match } from '@sourcegraph/shared/src/api/client/types/textDocument'
import { ExtensionCodeEditor } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { PanelViewData } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, useObservable, Tab, TabList, TabPanel, TabPanels, Tabs, Icon } from '@sourcegraph/wildcard'

import { registerPanelToolbarContributions } from './views/contributions'
import { EmptyPanelView } from './views/EmptyPanelView'
import { ExtensionsLoadingPanelView } from './views/ExtensionsLoadingView'
import { PanelView } from './views/PanelView'
import { ReferencesPanelFeedbackCta } from './views/ReferencesPanelFeedbackCta'

import styles from './TabbedPanelContent.module.scss'

interface TabbedPanelContentProps
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
     * Maximum number of results to show from locationProvider. If not set,
     * MAXIMUM_LOCATION_RESULTS will be used.
     */
    maxLocationResults?: number

    /**
     * The React element to render in the panel view.
     */
    reactElement?: React.ReactFragment

    // Should the content of the panel be put inside a wrapper container with padding or not.
    noWrapper?: boolean

    // Should the panel be shown for the given `#tab=<ID>` in the URL?
    matchesTabID?: (id: string) => boolean
}

/**
 * A tab and corresponding content to display in the panel.
 */
interface TabbedPanelItem {
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

    /** Callback that's triggered when the panel is selected */
    trackTabClick: () => void

    // Should the panel item be shown for the given `#tab=<ID>` in the URL?
    matchesTabID?: (id: string) => boolean
}

export type BuiltinTabbedPanelView = Omit<PanelViewWithComponent, 'component' | 'id'>

const builtinTabbedPanelViewProviders = new BehaviorSubject<
    Map<string, { id: string; provider: Observable<BuiltinTabbedPanelView | null> }>
>(new Map())

/**
 * BuiltinTabbedPanelView defines which BuiltinTabbedPanelViews will be available.
 */
export interface BuiltinTabbedPanelDefinition {
    id: string
    provider: Observable<BuiltinTabbedPanelView | null>
}
/**
 * React hook to add panel views from other components (panel views are typically
 * contributed by Sourcegraph extensions)
 */
export function useBuiltinTabbedPanelViews(builtinPanels: BuiltinTabbedPanelDefinition[]): void {
    useEffect(() => {
        for (const builtinPanel of builtinPanels) {
            builtinTabbedPanelViewProviders.value.set(builtinPanel.id, builtinPanel)
        }
        builtinTabbedPanelViewProviders.next(new Map([...builtinTabbedPanelViewProviders.value]))

        return () => {
            for (const builtinPanel of builtinPanels) {
                builtinTabbedPanelViewProviders.value.delete(builtinPanel.id)
            }
            builtinTabbedPanelViewProviders.next(new Map([...builtinTabbedPanelViewProviders.value]))
        }
    }, [builtinPanels])
}

/**
 * The panel, which is a tabbed component with contextual information. Components rendering the panel should
 * generally use ResizablePanel, not Panel.
 *
 * Other components can contribute panel items to the panel with the `useBuildinPanelViews` hook.
 */
export const TabbedPanelContent = React.memo<TabbedPanelContentProps>(props => {
    // Ensures that we don't show a misleading empty state when extensions haven't loaded yet.
    const areExtensionsReady = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(props.extensionsController.extHostAPI), [props.extensionsController])
    )
    const [redesignedEnabled] = useTemporarySetting('codeintel.referencePanel.redesign.enabled', false)
    const isExperimentalReferencePanelEnabled =
        (!isErrorLike(props.settingsCascade.final) &&
            props.settingsCascade.final?.experimentalFeatures?.coolCodeIntel === true) ||
        redesignedEnabled === true

    const [tabIndex, setTabIndex] = useState(0)
    const location = useLocation()
    const { hash, pathname, search } = location
    const history = useHistory()
    const handlePanelClose = useCallback(() => history.replace(pathname), [history, pathname])
    const [currentTabLabel, currentTabID] = hash.split('=')

    const builtinTabbedPanels: PanelViewWithComponent[] | undefined = useObservable(
        useMemo(
            () =>
                builtinTabbedPanelViewProviders.pipe(
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
                        combineLatest([
                            wrapRemoteObservable(extensionHostAPI.getPanelViews()),
                            wrapRemoteObservable(extensionHostAPI.getActiveViewComponentChanges()),
                        ]).pipe(
                            switchMap(async ([panelViews, viewer]) => {
                                if ((await viewer?.type) !== 'CodeEditor') {
                                    return undefined
                                }

                                const document = await (viewer as Remote<ExtensionCodeEditor>).document

                                return panelViews
                                    .filter(panelView =>
                                        panelView.selector !== null ? match(panelView.selector, document) : true
                                    )
                                    .filter(panelView =>
                                        // If we use the new reference panel we don't want to display additional
                                        // 'implementations_' panels
                                        isExperimentalReferencePanelEnabled
                                            ? !panelView.component?.locationProvider?.startsWith('implementations_')
                                            : true
                                    )
                                    .map((panelView: PanelViewWithComponent) => {
                                        const locationProviderID = panelView.component?.locationProvider
                                        const maxLocations = panelView.component?.maxLocationResults
                                        if (locationProviderID) {
                                            const panelViewWithProvider: PanelViewWithComponent = {
                                                ...panelView,
                                                maxLocationResults: maxLocations,
                                                locationProvider: wrapRemoteObservable(
                                                    extensionHostAPI.getActiveCodeEditorPosition()
                                                ).pipe(
                                                    switchMap(parameters => {
                                                        if (!parameters) {
                                                            return [{ isLoading: false, result: [] }]
                                                        }

                                                        return wrapRemoteObservable(
                                                            extensionHostAPI.getLocations(
                                                                locationProviderID,
                                                                parameters
                                                            )
                                                        )
                                                    })
                                                ),
                                            }
                                            return panelViewWithProvider
                                        }

                                        return panelView
                                    })
                            })
                        )
                    )
                ),
            [isExperimentalReferencePanelEnabled, props.extensionsController]
        )
    )

    const panelViews = useMemo(() => [...(builtinTabbedPanels || []), ...(extensionPanels || [])], [
        builtinTabbedPanels,
        extensionPanels,
    ])

    const trackTabClick = useCallback((label: string) => props.telemetryService.log(`ReferencePanelClicked${label}`), [
        props.telemetryService,
    ])

    const items = useMemo(
        () =>
            panelViews
                ? panelViews
                      .map(
                          (panelView): TabbedPanelItem => ({
                              label: panelView.title,
                              id: panelView.id,
                              priority: panelView.priority,
                              element: <PanelView {...props} panelView={panelView} location={location} />,
                              hasLocations: !!panelView.locationProvider,
                              trackTabClick: () => trackTabClick(panelView.title),
                              matchesTabID: panelView.matchesTabID,
                          })
                      )
                      .sort((a, b) => b.priority - a.priority)
                : [],
        [location, panelViews, props, trackTabClick]
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
        setTabIndex(
            items.findIndex(({ id, matchesTabID }) => (matchesTabID ? matchesTabID(currentTabID) : id === currentTabID))
        )
    }, [items, hash, currentTabID])

    if (!areExtensionsReady) {
        return <ExtensionsLoadingPanelView className={styles.panel} />
    }

    if (!items) {
        return <EmptyPanelView className={styles.panel} />
    }

    const activeTab: TabbedPanelItem | undefined = items[tabIndex]

    return (
        <Tabs className={styles.panel} index={tabIndex} onChange={handleActiveTab}>
            <TabList
                wrapperClassName={classNames(styles.panelHeader, 'sticky-top')}
                actions={
                    <div className="align-items-center d-flex">
                        {activeTab && (
                            <>
                                {(activeTab.id === 'def' ||
                                    activeTab.id === 'references' ||
                                    activeTab.id.startsWith('implementations_')) && <ReferencesPanelFeedbackCta />}
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
                            </>
                        )}
                        <Button
                            onClick={handlePanelClose}
                            variant="icon"
                            className={classNames('ml-2', styles.dismissButton)}
                            title="Close panel"
                            data-tooltip="Close panel"
                            data-placement="left"
                        >
                            <Icon role="img" as={CloseIcon} aria-hidden={true} />
                        </Button>
                    </div>
                }
            >
                {items.map(({ label, id, trackTabClick }, index) => (
                    <Tab key={id} index={index}>
                        <span className="tablist-wrapper--tab-label" onClick={trackTabClick} role="none">
                            {label}
                        </span>
                    </Tab>
                ))}
            </TabList>
            <TabPanels>
                {activeTab ? (
                    items.map(({ id, element }, index) => (
                        <TabPanel
                            index={index}
                            key={id}
                            className={styles.tabsContent}
                            data-testid="panel-tabs-content"
                        >
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
