import * as H from 'history'
import CancelIcon from 'mdi-react/CancelIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { ReplaySubject, Subject, Subscription, Unsubscribable } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'
import { Resizable } from '../components/Resizable'
import { Spacer, Tab, TabsWithURLViewStatePersistence } from '../components/Tabs'
import { eventLogger } from '../tracking/eventLogger'
import { parseHash } from '../util/url'

/**
 * A tab and corresponding content to display in the panel.
 */
export interface PanelItem extends Tab<string> {
    /**
     * Controls the relative order of panel items. The items are laid out from highest priority (at the beginning)
     * to lowest priority (at the end). The default is 0.
     */
    priority: number

    /** The content element to display when the tab is active. */
    element: React.ReactElement<any>
}

interface Props {
    isLightTheme: boolean
    location: H.Location
    history: H.History
}

interface State {
    /** Panel title. */
    title?: React.ReactFragment

    /** Panel items to display. */
    items: PanelItem[]
}

// TODO: Panel should be mounted whenever the blob page is shown.
// There's no reason to not have it rendered. We can just hide it. This will make the logic here a lot safer.

/**
 * The panel, which is a tabbed component with contextual information. Components rendering the panel should
 * generally use ResizablePanel, not Panel.
 *
 * Other components can contribute panel items to the panel.
 */
export class Panel extends React.PureComponent<Props, State> {
    private static forceUpdates = new Subject<void>()

    // We'll use ReplaySubject so everything will get emitted even if the panel isn't mounted
    private static newItems = new ReplaySubject<PanelItem>()
    private static newTitles = new ReplaySubject<React.ReactFragment>()
    private static itemsToRemove = new ReplaySubject<PanelItem>()
    private static titlesToRemove = new ReplaySubject<React.ReactFragment>()

    private subscriptions = new Subscription()

    public state: State = { items: [] }

    /**
     * Set the panel title. Do not call directly; use PanelTitlePortal instead.
     * @param fragment to set as the panel title
     */
    public static setTitle(fragment: React.ReactFragment | undefined): Unsubscribable {
        Panel.newTitles.next(fragment)
        return {
            unsubscribe: () => {
                Panel.titlesToRemove.next(fragment)
            },
        }
    }

    /**
     * Add an item to the panel. Do not call directly; use PanelItemPortal instead.
     * @param item to add to the header
     */
    public static addItem(item: PanelItem): Unsubscribable {
        Panel.newItems.next(item)
        return {
            unsubscribe: () => {
                Panel.itemsToRemove.next(item)
            },
        }
    }

    /**
     * Forces an update of items in the panel. Do not call directly; use PanelItemPortal instead.
     */
    public static forceUpdate(): void {
        this.forceUpdates.next()
    }

    public componentDidMount(): void {
        this.subscriptions.add(Panel.forceUpdates.subscribe(() => this.forceUpdate()))

        this.subscriptions.add(
            Panel.newItems.pipe(distinctUntilChanged()).subscribe(item => {
                this.setState(state => ({
                    items: state.items.concat(item).sort(byPriority),
                }))
            })
        )

        this.subscriptions.add(
            Panel.itemsToRemove.subscribe(item => {
                this.setState(state => ({
                    items: state.items.filter(i => i !== item),
                }))
            })
        )

        this.subscriptions.add(
            Panel.newTitles.pipe(distinctUntilChanged()).subscribe(title => {
                this.setState({
                    title,
                })
            })
        )

        this.subscriptions.add(
            Panel.titlesToRemove.subscribe(title => {
                if (this.state.title === title) {
                    this.setState({ title: undefined })
                }
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const closeButton = (
            <button
                onClick={this.onDismiss}
                className="btn btn-icon panel__header-icon tab_bar__close-button tab-bar__end-fragment-other-element"
                data-tooltip="Close"
            >
                <CloseIcon className="icon-inline" />
            </button>
        )

        const hasTabs = this.state.items.length > 0

        return (
            <div className="panel">
                {(!hasTabs || this.state.title) && (
                    <header className="panel__header">
                        <div className="panel__header-title">{this.state.title}</div>
                        {closeButton}
                    </header>
                )}
                {hasTabs ? (
                    <TabsWithURLViewStatePersistence
                        tabs={this.state.items || []}
                        tabBarEndFragment={
                            <>
                                <Spacer />
                                {!this.state.title && closeButton}
                            </>
                        }
                        className="panel__tabs"
                        tabClassName="tab-bar__tab--h5like"
                        onSelectTab={this.onSelectTab}
                        location={this.props.location}
                    >
                        {this.state.items &&
                            this.state.items.map(({ id, element }) => React.cloneElement(element, { key: id }))}
                    </TabsWithURLViewStatePersistence>
                ) : (
                    <div className="panel__empty">
                        <CancelIcon className="icon-inline" /> Nothing to show here
                    </div>
                )}
            </div>
        )
    }

    private onSelectTab = (tab: string): void => eventLogger.log('PanelTabActivated', { tab })

    private onDismiss = (): void =>
        this.props.history.push(TabsWithURLViewStatePersistence.urlForTabID(this.props.location, null))
}

function byPriority(a: { priority: number }, b: { priority: number }): number {
    return b.priority - a.priority
}

/** A wrapper around Panel that makes it resizable. */
export const ResizablePanel: React.SFC<Props> = props =>
    !!parseHash(props.location.hash).viewState ? (
        <Resizable
            className="panel--resizable"
            handlePosition="top"
            defaultSize={350}
            storageKey="panel-size"
            element={<Panel {...props} />}
        />
    ) : null
