import classNames from 'classnames'
import React from 'react'

import { ActionsNavItems, ActionsNavItemsProps } from '@sourcegraph/shared/src/actions/ActionsNavItems'
import {
    CommandListPopoverButton,
    CommandListPopoverButtonProps,
} from '@sourcegraph/shared/src/commandPalette/CommandList'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

// Components from shared with web-styling class names applied

export { WebHoverOverlay } from './WebHoverOverlay'

export const WebCommandListPopoverButton: React.FunctionComponent<CommandListPopoverButtonProps> = props => {
    const [isRedesignEnabled] = useRedesignToggle()
    return (
        <CommandListPopoverButton
            {...props}
            buttonClassName="btn btn-link"
            popoverClassName={classNames('popover', isRedesignEnabled && 'border-0')}
            popoverInnerClassName="rounded overflow-hidden"
            formClassName={classNames('form', isRedesignEnabled && 'p-2 bg-1 border-bottom')}
            inputClassName="form-control px-2 py-1"
            listClassName={classNames('list-group list-group-flush list-unstyled', isRedesignEnabled && 'pt-1')}
            actionItemClassName={classNames(
                'list-group-item list-group-item-action',
                isRedesignEnabled ? 'p-2 border-0' : 'px-2'
            )}
            selectedActionItemClassName="active border-primary"
            noResultsClassName="list-group-item text-muted"
        />
    )
}

WebCommandListPopoverButton.displayName = 'WebCommandListPopoverButton'

export const WebActionsNavItems: React.FunctionComponent<ActionsNavItemsProps> = ({
    listClass,
    listItemClass,
    actionItemClass,
    actionItemIconClass,
    ...props
}) => (
    <ActionsNavItems
        {...props}
        listClass={classNames(listClass, 'nav')}
        listItemClass={classNames(listItemClass, 'nav-item')}
        actionItemClass={classNames(actionItemClass, 'nav-link')}
        actionItemIconClass={classNames(actionItemIconClass, 'icon-inline-md')}
    />
)
WebActionsNavItems.displayName = 'WebActionsNavItems'
