import * as React from 'react'
import { Subject, Subscription, combineLatest } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { ContributionScope, Context } from '../api/client/context/context'
import { getContributedActionItems } from '../contributions/contributions'
import { TelemetryProps } from '../telemetry/telemetryService'
import { ActionItem, ActionItemProps } from './ActionItem'
import { ActionsState } from './actions'
import { ActionsProps } from './ActionsContainer'

export interface ActionNavItemsClassProps {
    /**
     * CSS class name for one action item (`<button>` or `<a>`)
     */
    actionItemClass?: string

    /**
     * Additional CSS class name when the action item is a toogle in its enabled state.
     */
    actionItemPressedClass?: string

    actionItemIconClass?: string

    /**
     * CSS class name for each `<li>` element wrapping the action item.
     */
    listItemClass?: string
}

export interface ActionsNavItemsProps
    extends ActionsProps,
        ActionNavItemsClassProps,
        TelemetryProps,
        Pick<ActionItemProps, 'showLoadingSpinnerDuringExecution'> {
    /**
     * If true, it renders a `<ul className="nav">...</ul>` around the items. If there are no items, it renders `null`.
     *
     * If falsey (the default behavior), it emits a fragment of just the `<li>`s.
     */
    wrapInList?: boolean
    /**
     * Only applied if `wrapInList` is `true`
     */

    listClass?: string
}

/**
 * Renders the actions as a fragment of <li class="nav-item"> elements, for use in a Bootstrap <ul
 * class="nav"> or <ul class="navbar-nav">.
 */
export class ActionsNavItems extends React.PureComponent<ActionsNavItemsProps, ActionsState> {
    public state: ActionsState = {}

    private scopeChanges = new Subject<ContributionScope | undefined>()
    private extraContextChanges = new Subject<Context<any> | undefined>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        that.subscriptions.add(
            combineLatest([that.scopeChanges, that.extraContextChanges])
                .pipe(
                    switchMap(([scope, extraContext]) =>
                        that.props.extensionsController.services.contribution.getContributions(scope, extraContext)
                    )
                )
                .subscribe(contributions => that.setState({ contributions }))
        )
        that.scopeChanges.next(that.props.scope)
        that.extraContextChanges.next(that.props.extraContext)
    }

    public componentDidUpdate(prevProps: ActionsProps): void {
        if (prevProps.scope !== that.props.scope) {
            that.scopeChanges.next(that.props.scope)
        }
        if (prevProps.extraContext !== that.props.extraContext) {
            that.extraContextChanges.next(that.props.extraContext)
        }
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | React.ReactFragment | null {
        if (!that.state.contributions) {
            return null // loading
        }

        const actionItems = getContributedActionItems(that.state.contributions, that.props.menu).map((item, i) => (
            <React.Fragment key={item.action.id}>
                {' '}
                <li className={that.props.listItemClass}>
                    <ActionItem
                        key={item.action.id}
                        {...item}
                        {...that.props}
                        variant="actionItem"
                        iconClassName={that.props.actionItemIconClass}
                        className={that.props.actionItemClass}
                        pressedClassName={that.props.actionItemPressedClass}
                    />
                </li>
            </React.Fragment>
        ))
        if (that.props.wrapInList) {
            return actionItems.length > 0 ? <ul className={that.props.listClass}>{actionItems}</ul> : null
        }
        return <>{actionItems}</>
    }
}
