import classNames from 'classnames'
import React from 'react'
import { ActionsNavItems, ActionsNavItemsProps } from '../../../shared/src/actions/ActionsNavItems'
import { CommandListPopoverButton, CommandListPopoverButtonProps } from '../../../shared/src/commandPalette/CommandList'
import { HoverOverlay, HoverOverlayProps } from '../../../shared/src/hover/HoverOverlay'

// Components from shared with web-styling class names applied

export const WebHoverOverlay: React.FunctionComponent<HoverOverlayProps> = props => (
    <HoverOverlay actionItemClassName="btn btn-secondary" {...props} />
)
WebHoverOverlay.displayName = 'WebHoverOverlay'

export const WebCommandListPopoverButton: React.FunctionComponent<CommandListPopoverButtonProps> = props => (
    <CommandListPopoverButton
        {...props}
        popoverClassName="rounded"
        formClassName="form"
        inputClassName="form-control px-2 py-1 rounded-0"
        listClassName="list-group list-group-flush list-unstyled"
        actionItemClassName="list-group-item list-group-item-action px-2"
        selectedActionItemClassName="active border-primary"
        noResultsClassName="list-group-item text-muted bg-striped-secondary"
    />
)
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
        actionItemIconClass={classNames(actionItemIconClass, 'icon-inline')}
    />
)
WebActionsNavItems.displayName = 'WebActionsNavItems'
