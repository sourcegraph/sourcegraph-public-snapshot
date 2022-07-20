import React, { useCallback, useState } from 'react'

import { mdiClose } from '@mdi/js'
import { Shortcut, ModifierKey, Key } from '@slimsag/react-shortcuts'

import { Button, Modal, Icon, H4 } from '@sourcegraph/wildcard'

import { KeyboardShortcut } from '../keyboardShortcuts'

import { KeyboardShortcutsProps, KEYBOARD_SHORTCUTS } from './keyboardShortcuts'

import styles from './KeyboardShortcutsHelp.module.scss'

interface Props {
    isOpen?: boolean
    onDismiss: () => void
}

/**
 * Keyboard shortcuts that are implemented in a legacy way, not using the central keyboard shortcuts
 * registry. These are shown in the help modal.
 * // TODO: CHECK IF USED
 */
const LEGACY_KEYBOARD_SHORTCUTS: KeyboardShortcut[] = [
    {
        id: 'canonicalURL',
        title: 'Expand URL to its canonical form (on file or tree page)',
        keybindings: [{ ordered: ['y'] }],
    },
]

const KEY_TO_NAMES: { [P in Key | ModifierKey | string]?: string } = {
    Meta: 'Cmd',
    Control: 'Ctrl',
    '†': 't',
}

const MODAL_LABEL_ID = 'keyboard-shortcuts-help-modal-title'

export const KeyboardShortcutsHelp: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    isOpen,
    onDismiss,
}) => {
    return (
        <Modal
            position="center"
            isOpen={isOpen}
            onDismiss={onDismiss}
            aria-labelledby={MODAL_LABEL_ID}
            containerClassName={styles.modalContainer}
        >
            <div className={styles.modalHeader}>
                <H4 id={MODAL_LABEL_ID}>Keyboard shortcuts</H4>
                <Button variant="icon" aria-label="Close" onClick={onDismiss}>
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <div>
                <ul className="list-group list-group-flush">
                    {Object.values(KEYBOARD_SHORTCUTS)
                        .filter(({ hideInHelp }) => !hideInHelp)
                        .map(({ title, keybindings }, index) => (
                            <li
                                key={index}
                                className="list-group-item d-flex align-items-center justify-content-between"
                            >
                                {title}
                                <span>
                                    {keybindings.map((keybinding, index) => (
                                        <span key={index}>
                                            {index !== 0 && ' or '}
                                            {[...(keybinding.held || []), ...keybinding.ordered].map((key, index) => (
                                                <kbd key={index}>{KEY_TO_NAMES[key] ?? key}</kbd>
                                            ))}
                                        </span>
                                    ))}
                                </span>
                            </li>
                        ))}
                </ul>
            </div>
        </Modal>
    )
}
