import * as H from 'history'
import marked from 'marked'
import CancelIcon from 'mdi-react/CancelIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { ContributableViewContainer } from '../../../shared/src/api/protocol/contribution'
import { PanelView } from '../../../shared/src/api/protocol/plainTypes'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { Markdown } from '../components/Markdown'
import { Resizable } from '../components/Resizable'
import { Spacer, Tab, TabsWithURLViewStatePersistence } from '../components/Tabs'
import { createLinkClickHandler } from '../util/linkClickHandler'

const EmptyPanelView: React.FunctionComponent = () => (
    <div className="panel__empty">
        <CancelIcon className="icon-inline" /> Nothing to show here
    </div>
)

interface Props extends ExtensionsControllerProps {
    location: H.Location
    history: H.History
}

interface State {
    /** Panel views contributed by extensions. */
    panelViews?: (PanelView & { id: string })[] | null
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
                .pipe(map(panelViews => ({ panelViews })))
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const items = this.state.panelViews
            ? this.state.panelViews
                  .map(
                      panelView =>
                          ({
                              label: panelView.title,
                              id: panelView.id,
                              priority: 0, // TODO!(sqs)
                              element:
                                  panelView.component ||
                                  ((
                                      <div className="p-2" onClick={createLinkClickHandler(this.props.history)}>
                                          <Markdown dangerousInnerHTML={marked(panelView.content)} />
                                      </div>
                                  ) || <EmptyPanelView />),
                          } as PanelItem)
                  )
                  .sort(byPriority)
            : []

        const hasTabs = items.length > 0

        return (
            <div className="panel">
                {hasTabs ? (
                    <TabsWithURLViewStatePersistence
                        tabs={items || []}
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
// TODO!(sqs): show this selectively
export const ResizablePanel: React.FunctionComponent<Props> = props => (
    <Resizable
        className="panel--resizable"
        handlePosition="top"
        defaultSize={350}
        storageKey="panel-size"
        element={<Panel {...props} />}
    />
)
