import classNames from 'classnames'
import React from 'react'
import { ActionsNavItems, ActionsNavItemsProps } from '../../../shared/src/actions/ActionsNavItems'
import { CommandListPopoverButton, CommandListPopoverButtonProps } from '../../../shared/src/commandPalette/CommandList'
import {
    EditorCompletionWidget,
    EditorCompletionWidgetProps,
} from '../../../shared/src/components/completion/EditorCompletionWidget'
import { HoverOverlay, HoverOverlayProps } from '../../../shared/src/hover/HoverOverlay'

// Components from shared with web-styling class names applied

export const WebHoverOverlay: React.FunctionComponent<HoverOverlayProps<never>> = props => (
    <HoverOverlay
        {...props}
        className="card"
        iconClassName="icon-inline"
        closeButtonClassName="btn btn-icon"
        actionItemClassName="btn btn-secondary"
        infoAlertClassName="alert alert-info"
        errorAlertClassName="alert alert-danger"
    />
)
WebHoverOverlay.displayName = 'WebHoverOverlay'

export const WebCommandListPopoverButton: React.FunctionComponent<CommandListPopoverButtonProps> = props => (
    <CommandListPopoverButton
        {...props}
        buttonClassName="btn btn-link"
        popoverClassName="popover"
        popoverInnerClassName="border rounded overflow-hidden"
        formClassName="form"
        inputClassName="form-control px-2 py-1 rounded-0"
        listClassName="list-group list-group-flush list-unstyled"
        actionItemClassName="list-group-item list-group-item-action px-2"
        selectedActionItemClassName="active border-primary"
        noResultsClassName="list-group-item text-muted"
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

const completionWidgetListItemClassName = 'completion-widget-dropdown__item d-flex align-items-center p-2'

export const WebEditorCompletionWidget: React.FunctionComponent<EditorCompletionWidgetProps> = props => (
    <EditorCompletionWidget
        {...props}
        listClassName="completion-widget-dropdown d-block list-unstyled rounded p-0 m-0 mt-3"
        listItemClassName={completionWidgetListItemClassName}
        selectedListItemClassName="completion-widget-dropdown__item--selected bg-primary"
        loadingClassName={completionWidgetListItemClassName}
        noResultsClassName={completionWidgetListItemClassName}
    />
)
WebEditorCompletionWidget.displayName = 'WebEditorCompletionWidget'
