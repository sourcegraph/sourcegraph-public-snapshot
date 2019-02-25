import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { ContributionScope } from '../api/client/context/context'
import { getContributedActionItems } from '../contributions/contributions'
import { ActionItem } from './ActionItem'
import { ActionsState } from './actions'
import { ActionsProps } from './ActionsContainer'

interface Props extends ActionsProps {
    /**
     * If true, it renders a <ul className="nav">...</ul> around the items. If there are no items, it renders null.
     *
     * If falsey (the default behavior), it emits a fragment of just the <li>s.
     */
    wrapInList?: boolean

    listClass?: string
    actionItemClass?: string
    actionItemPressedClass?: string
    listItemClass?: string
}

/**
 * Renders the actions as a fragment of <li class="nav-item"> elements, for use in a Bootstrap <ul
 * class="nav"> or <ul class="navbar-nav">.
 */
export class ActionsNavItems extends React.PureComponent<Props, ActionsState> {
    public state: ActionsState = {}

    private scopeChanges = new Subject<ContributionScope | undefined>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.scopeChanges
                .pipe(switchMap(scope => this.props.extensionsController.services.contribution.getContributions(scope)))
                .subscribe(contributions => this.setState({ contributions }))
        )
        this.scopeChanges.next(this.props.scope)
    }

    public componentDidUpdate(prevProps: ActionsProps): void {
        if (prevProps.scope !== this.props.scope) {
            this.scopeChanges.next(this.props.scope)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | React.ReactFragment | null {
        if (!this.state.contributions) {
            return null // loading
        }

        const actionItems = getContributedActionItems(this.state.contributions, this.props.menu).map((item, i) => (
            <li key={i} className={this.props.listItemClass || 'nav-item'}>
                <ActionItem
                    key={i}
                    {...item}
                    variant="actionItem"
                    extensionsController={this.props.extensionsController}
                    platformContext={this.props.platformContext}
                    className={this.props.actionItemClass}
                    pressedClassName={this.props.actionItemPressedClass}
                    location={this.props.location}
                />
            </li>
        ))
        if (this.props.wrapInList) {
            return actionItems.length > 0 ? (
                <ul className={`nav ${this.props.listClass || ''}`}>{actionItems}</ul>
            ) : null
        }
        return <>{actionItems}</>
    }
}
