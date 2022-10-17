import React from 'react'

import classNames from 'classnames'

import {
    CommandListPopoverButton,
    CommandListPopoverButtonProps,
} from '@sourcegraph/shared/src/commandPalette/CommandList'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { Button } from '@sourcegraph/wildcard'

import styles from './WebCommandListPopoverButton.module.scss'

export const WebCommandListPopoverButton: React.FunctionComponent<
    React.PropsWithChildren<CommandListPopoverButtonProps>
> = props => {
    const showCommandPaletteShortcut = useKeyboardShortcut('commandPalette')

    return (
        <CommandListPopoverButton
            {...props}
            as={Button}
            variant="link"
            buttonClassName={classNames('m-0 p-0', styles.button)}
            formClassName="form p-2 bg-1 border-bottom"
            inputClassName="form-control px-2 py-1"
            listClassName="list-group list-group-flush list-unstyled pt-1"
            actionItemClassName={classNames('list-group-item list-group-item-action p-2 border-0', styles.actionItem)}
            selectedActionItemClassName="active border-primary"
            noResultsClassName="list-group-item text-muted"
            keyboardShortcutForShow={showCommandPaletteShortcut}
        />
    )
}

WebCommandListPopoverButton.displayName = 'WebCommandListPopoverButton'
