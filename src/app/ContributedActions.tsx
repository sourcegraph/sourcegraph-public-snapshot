import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { sortBy } from 'lodash'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ExtensionsProps } from '../context'
import { Settings } from '../copypasta'
import { CXPControllerProps } from '../cxp/controller'
import { ConfigurationSubject } from '../settings'
import { ContributedActionItem, ContributedActionItemProps } from './ContributedActionItem'

interface ContributedActionsProps<S extends ConfigurationSubject, C = Settings>
    extends CXPControllerProps,
        ExtensionsProps<S, C> {
    menu: ContributableMenu
}

interface ContributedActionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}

/**
 * Renders the contributions for a menu as a fragment of <li class="nav-item"> elements, for use in a Bootstrap <ul
 * class="nav"> or <ul class="navbar-nav">.
 */
export class ContributedActionsNavItems<S extends ConfigurationSubject, C = Settings> extends React.PureComponent<
    ContributedActionsProps<S, C>,
    ContributedActionsState
> {
    public state: ContributedActionsState = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.cxpController.registries.contribution.contributions.subscribe(contributions =>
                this.setState({ contributions })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.contributions) {
            return null // loading
        }

        return (
            <>
                {getContributedActionItems(this.state.contributions, this.props.menu).map((item, i) => (
                    <li key={i} className="nav-item">
                        <ContributedActionItem
                            key={i}
                            {...item}
                            variant="toolbarItem"
                            cxpController={this.props.cxpController}
                            extensions={this.props.extensions}
                        />
                    </li>
                ))}
            </>
        )
    }
}

interface ContributedActionsContainerProps<S extends ConfigurationSubject, C = Settings>
    extends ContributedActionsProps<S, C> {
    /**
     * Called with the array of contributed items to produce the rendered component. If not set, uses a default
     * render function that renders a <ContributedActionItem> for each item.
     */
    render?: (items: ContributedActionItemProps[]) => React.ReactElement<any>

    /**
     * If set, it is rendered when there are no contributed items for this menu. Use null to render nothing when
     * empty.
     */
    empty?: React.ReactElement<any> | null
}

interface ContributedActionsContainerState extends ContributedActionsState {}

/** Displays the contributions actions for a menu, with a wrapper and/or empty element. */
export class ContributedActionsContainer<S extends ConfigurationSubject, C = Settings> extends React.PureComponent<
    ContributedActionsContainerProps<S, C>,
    ContributedActionsContainerState
> {
    public state: ContributedActionsState = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.cxpController.registries.contribution.contributions.subscribe(contributions =>
                this.setState({ contributions })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.contributions) {
            return null // loading
        }

        const items = getContributedActionItems(this.state.contributions, this.props.menu)
        if (this.props.empty !== undefined && items.length === 0) {
            return this.props.empty
        }

        const render = this.props.render || this.defaultRenderItems
        return render(items)
    }

    private defaultRenderItems = (items: ContributedActionItemProps[]): JSX.Element | null => (
        <>
            {items.map((item, i) => (
                <ContributedActionItem
                    key={i}
                    {...item}
                    cxpController={this.props.cxpController}
                    extensions={this.props.extensions}
                />
            ))}
        </>
    )
}

/** Collect all command contrbutions for the menu. */
export function getContributedActionItems(
    contributions: Contributions,
    menu: ContributableMenu
): ContributedActionItemProps[] {
    const allItems: ContributedActionItemProps[] = []
    const menuItems = contributions.menus && contributions.menus[menu]
    if (menuItems) {
        for (const { command: commandID } of menuItems) {
            const command = contributions.commands && contributions.commands.find(c => c.command === commandID)
            if (command) {
                allItems.push({ contribution: command })
            }
        }
    }
    return sortBy(allItems, (item: ContributedActionItemProps) => item.contribution.command)
}
