import classNames from 'classnames'
import React from 'react'

import {
    CommandListPopoverButton,
    CommandListPopoverButtonProps,
} from '@sourcegraph/shared/src/commandPalette/CommandList'

import styles from './WebCommandListPopoverButton.module.scss'

// Components from shared with web-styling class names applied
export { WebHoverOverlay } from './WebHoverOverlay'

export const WebCommandListPopoverButton: React.FunctionComponent<CommandListPopoverButtonProps> = props => (
    <CommandListPopoverButton
        {...props}
        variant="link"
        buttonClassName={classNames('m-0 p-0', styles.button)}
        popoverClassName="popover border-0"
        popoverInnerClassName="rounded overflow-hidden"
        formClassName="form p-2 bg-1 border-bottom"
        inputClassName="form-control px-2 py-1"
        listClassName="list-group list-group-flush list-unstyled pt-1"
        actionItemClassName="list-group-item list-group-item-action p-2 border-0"
        selectedActionItemClassName="active border-primary"
        noResultsClassName="list-group-item text-muted"
    />
)

WebCommandListPopoverButton.displayName = 'WebCommandListPopoverButton'
