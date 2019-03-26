import * as sgtypes from '@sourcegraph/extension-api-types'
import * as H from 'history'
import { merge } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { Observable, of, Subscription } from 'rxjs'
import { catchError, endWith, map, mergeMap, startWith, switchMap, tap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ContributableMenu, ContributableViewContainer } from '../../../shared/src/api/protocol/contribution'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { ActionsNavItems } from '../actions/ActionsNavItems'
import { PanelViewWithComponent, ViewProviderRegistrationOptions } from '../api/client/services/view'
import { ActivationProps } from '../components/activation/Activation'
import { FetchFileCtx } from '../components/CodeExcerpt'
import { Resizable } from '../components/Resizable'
import { Spacer, Tab, TabsWithURLViewStatePersistence } from '../components/Tabs'
import { PlatformContextProps } from '../platform/context'
import { SettingsCascadeProps } from '../settings/settings'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { combineLatestOrDefault } from '../util/rxjs/combineLatestOrDefault'
import { EmptyPanelView } from './views/EmptyPanelView'
import { PanelView } from './views/PanelView'

interface Props extends ExtensionsControllerProps, PlatformContextProps, SettingsCascadeProps, ActivationProps {
    location: H.Location
    history: H.History
    repoName?: string
    isLightTheme: boolean
    fetchHighlightedFileLines: (ctx: FetchFileCtx, force?: boolean) => Observable<string[]>
}

interface T extends Pick<sourcegraph.PanelView, 'title' | 'content' | 'priority'> {
    id: string
    locationsOrCustom:
        | { locations: { results?: Location[]; loading: boolean } | ErrorLike }
        | { custom: React.ReactFragment }
}

interface State {
    /** Panel views contributed by extensions. */
    panelViews?: T[]
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
    element: React.ReactElement<any>

    /**
     * Whether this panel contains a list of locations (from a location provider). This value is
     * exposed to contributions as `panel.activeView.hasLocations`. It is true if there is a
     * location provider (even if the result set is empty).
     */
    hasLocations?: boolean
}

function munch(
    props: any,
    a: Observable<sgtypes.Location[] | null>
): Observable<{ locations: { results?: Location[]; loading: boolean } | ErrorLike }> {
    return a.pipe(
        catchError((error): [ErrorLike] => [asError(error)]),
        map(result => ({
            locations: isErrorLike(result) ? result : { results: result || [], loading: true },
        })),
        startWith<any>({
            locations: { loading: true },
        }),
        tap(({ locations }) => {
            props.extensionsController.services.context.data.next({
                ...props.extensionsController.services.context.data.value,
                'panel.locations.hasResults':
                    locations && !isErrorLike(locations) && !!locations.results && locations.results.length > 0,
            })
        }),
        endWith({ locations: { loading: false } })
    )
}

/**
 * The panel, which is a tabbed component with contextual information. Components rendering the panel should
 * generally use ResizablePanel, not Panel.
 *
 * Other components can contribute panel items to the panel.
 */
export class Panel extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.extensionsController.services.views
                .getViews(ContributableViewContainer.Panel)
                .pipe(
                    switchMap((panelViews: (PanelViewWithComponent & ViewProviderRegistrationOptions)[]) =>
                        combineLatestOrDefault(
                            // must emit immedyetleh
                            panelViews.map<Observable<T>>(x =>
                                'locations' in x.locationsOrCustom
                                    ? x.locationsOrCustom.locations.pipe(
                                          mergeMap(z => munch(this.props, z)),
                                          map(f => ({ ...x, locationsOrCustom: f }))
                                      )
                                    : of({ ...x, locationsOrCustom: { custom: x.locationsOrCustom.custom } })
                            )
                        )
                    ),
                    map(v => ({ panelViews: v }))
                )
                .subscribe(
                    stateUpdate => {
                        this.setState(old => merge({}, old, stateUpdate))
                    },
                    error => console.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const items = this.state.panelViews
            ? this.state.panelViews
                  .map(
                      (panelView): PanelItem => ({
                          label: panelView.title,
                          id: panelView.id,
                          priority: panelView.priority,
                          // TypeScript doesn't know theses two panelViews have the same type
                          element: <PanelView {...this.props} panelView={panelView as any} />,
                          hasLocations: 'locations' in panelView.locationsOrCustom,
                      })
                  )
                  .sort(byPriority)
            : []

        const hasTabs = items.length > 0
        const activePanelViewID = TabsWithURLViewStatePersistence.readFromURL(this.props.location, items)
        const activePanelView = items.find(item => item.id === activePanelViewID)

        return (
            <div className="panel">
                {hasTabs ? (
                    <TabsWithURLViewStatePersistence
                        tabs={items}
                        tabBarEndFragment={
                            <>
                                <Spacer />
                                <button
                                    onClick={this.onDismiss}
                                    className="btn btn-icon tab-bar__end-fragment-other-element"
                                    data-tooltip="Close"
                                >
                                    <CloseIcon className="icon-inline" />
                                </button>
                            </>
                        }
                        toolbarFragment={
                            <ActionsNavItems
                                {...this.props}
                                listClass="w-100 justify-content-end"
                                actionItemClass="nav-link"
                                menu={ContributableMenu.PanelToolbar}
                                scope={
                                    activePanelViewID !== undefined
                                        ? {
                                              type: 'panelView',
                                              id: activePanelViewID,
                                              hasLocations: Boolean(activePanelView && activePanelView.hasLocations),
                                          }
                                        : undefined
                                }
                                wrapInList={true}
                            />
                        }
                        className="panel__tabs"
                        tabClassName="tab-bar__tab--h5like"
                        location={this.props.location}
                    >
                        {items && items.map(({ id, element }) => React.cloneElement(element, { key: id }))}
                    </TabsWithURLViewStatePersistence>
                ) : (
                    <EmptyPanelView />
                )}
            </div>
        )
    }

    private onDismiss = (): void =>
        this.props.history.push(TabsWithURLViewStatePersistence.urlForTabID(this.props.location, null))
}

function byPriority(a: { priority: number }, b: { priority: number }): number {
    return b.priority - a.priority
}

/** A wrapper around Panel that makes it resizable. */
export const ResizablePanel: React.FunctionComponent<Props> = props => (
    <Resizable
        className="panel--resizable"
        handlePosition="top"
        defaultSize={350}
        storageKey="panel-size"
        element={<Panel {...props} />}
    />
)
