import React from 'react'

import { mdiClose } from '@mdi/js'
import { ModifierKey, Key } from '@slimsag/react-shortcuts'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { KEYBOARD_SHORTCUTS } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, Modal, Icon, H4, Label } from '@sourcegraph/wildcard'

import styles from './KeyboardShortcutsHelp.module.scss'

interface Props {
    isOpen?: boolean
    onDismiss: () => void
}

/**
 * Keyboard shortcuts that are implemented in a legacy way, not using the central keyboard shortcuts
 * registry. These are shown in the help modal.
 */
const LEGACY_KEYBOARD_SHORTCUTS: Record<string, KeyboardShortcut> = {
    canonicalURL: {
        title: 'Expand URL to its canonical form (on file or tree page)',
        keybindings: [{ ordered: ['y'] }],
    },
}

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
    const [characterKeyShortcutsEnabled, setCharacterKeyShortcutsEnabled] = useTemporarySetting(
        'characterKeyShortcuts.enabled',
        true
    )

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
                    {Object.values({ ...KEYBOARD_SHORTCUTS, ...LEGACY_KEYBOARD_SHORTCUTS })
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
                <Label className={styles.modalFooter}>
                    <Toggle
                        value={characterKeyShortcutsEnabled}
                        onToggle={() => setCharacterKeyShortcutsEnabled(previous => !previous)}
                        title="Toggle character key shortcuts"
                        className="mr-2"
                    />
                    Character key shortcuts {characterKeyShortcutsEnabled ? 'enabled ' : 'disabled'}
                </Label>
            </div>
        </Modal>
    )
}
