import React, { useMemo } from 'react'
import { from } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { getContributedActionItems } from '../contributions/contributions'
import { TelemetryProps } from '../telemetry/telemetryService'
import { ActionItem, ActionItemProps } from './ActionItem'
import { ActionsProps } from './ActionsContainer'
import classNames from 'classnames'
import { useObservable } from '../util/useObservable'
import { wrapRemoteObservable } from '../api/client/api/common'

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
export const ActionsNavItems: React.FunctionComponent<ActionsNavItemsProps> = props => {
    const { scope, extraContext, extensionsController, menu, wrapInList } = props

    const contributions = useObservable(
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI =>
                        wrapRemoteObservable(extensionHostAPI.getContributions(scope, extraContext))
                    )
                ),
            [scope, extraContext, extensionsController.extHostAPI]
        )
    )

    if (!contributions) {
        return null // loading
    }

    const actionItems = getContributedActionItems(contributions, menu).map(item => (
        <React.Fragment key={item.action.id}>
            {' '}
            <li className={props.listItemClass}>
                <ActionItem
                    key={item.action.id}
                    {...item}
                    {...props}
                    variant="actionItem"
                    iconClassName={props.actionItemIconClass}
                    className={classNames('actions-nav-items__action-item', props.actionItemClass)}
                    pressedClassName={props.actionItemPressedClass}
                />
            </li>
        </React.Fragment>
    ))

    if (wrapInList) {
        return actionItems.length > 0 ? <ul className={props.listClass}>{actionItems}</ul> : null
    }
    return <>{actionItems}</>
}
