import React, { useCallback, useState } from 'react'

import { Shortcut, ModifierKey, Key } from '@slimsag/react-shortcuts'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Modal, Icon, Typography } from '@sourcegraph/wildcard'

import { KeyboardShortcut } from '../keyboardShortcuts'

import { KeyboardShortcutsProps } from './keyboardShortcuts'

import styles from './KeyboardShortcutsHelp.module.scss'

interface Props extends KeyboardShortcutsProps {
    /** The keyboard shortcut to show this modal. */
    keyboardShortcutForShow: KeyboardShortcut
}

/**
 * Keyboard shortcuts that are implemented in a legacy way, not using the central keyboard shortcuts
 * registry. These are shown in the help modal.
 */
const LEGACY_KEYBOARD_SHORTCUTS: KeyboardShortcut[] = [
    {
        id: 'canonicalURL',
        title: 'Expand URL to its canonical form (on file or tree page)',
        keybindings: [{ ordered: ['y'] }],
    },
]

const KEY_TO_NAMES: { [P in Key | ModifierKey]?: string } = {
    Meta: 'Cmd',
    Control: 'Ctrl',
}

const MODAL_LABEL_ID = 'keyboard-shortcuts-help-modal-title'

export const KeyboardShortcutsHelp: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    keyboardShortcutForShow,
    keyboardShortcuts,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    return (
        <>
            {keyboardShortcutForShow.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={toggleIsOpen} />
            ))}
            <Modal
                position="center"
                isOpen={isOpen}
                onDismiss={toggleIsOpen}
                aria-labelledby={MODAL_LABEL_ID}
                containerClassName={styles.modalContainer}
            >
                <div className={styles.modalHeader}>
                    <Typography.H4 id={MODAL_LABEL_ID}>Keyboard shortcuts</Typography.H4>
                    <Button variant="icon" aria-label="Close" onClick={toggleIsOpen}>
                        <Icon role="img" as={CloseIcon} aria-hidden={true} />
                    </Button>
                </div>
                <div>
                    <ul className="list-group list-group-flush">
                        {[...keyboardShortcuts, ...LEGACY_KEYBOARD_SHORTCUTS]
                            .filter(({ hideInHelp }) => !hideInHelp)
                            .map(({ id, title, keybindings }) => (
                                <li
                                    key={id}
                                    className="list-group-item d-flex align-items-center justify-content-between"
                                >
                                    {title}
                                    <span>
                                        {keybindings.map((keybinding, index) => (
                                            <span key={index}>
                                                {index !== 0 && ' or '}
                                                {[...(keybinding.held || []), ...keybinding.ordered].map(
                                                    (key, index) => (
                                                        <kbd key={index}>{KEY_TO_NAMES[key] ?? key}</kbd>
                                                    )
                                                )}
                                            </span>
                                        ))}
                                    </span>
                                </li>
                            ))}
                    </ul>
                </div>
            </Modal>
        </>
    )
}
