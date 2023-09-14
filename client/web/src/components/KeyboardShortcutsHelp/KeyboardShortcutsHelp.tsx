import React from 'react'

import { mdiClose } from '@mdi/js'
import { omit } from 'lodash'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { isMacPlatform } from '@sourcegraph/common'
import { type Keybinding, type KeyboardShortcut, shortcutDisplayName } from '@sourcegraph/shared/src/keyboardShortcuts'
import {
    KEYBOARD_SHORTCUTS,
    EXPERIMENTAL_BLOB_PAGE_SHORTCUTS,
} from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, Modal, Icon, H4, Label } from '@sourcegraph/wildcard'

import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'

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
const MODAL_LABEL_ID = 'keyboard-shortcuts-help-modal-title'

export const KeyboardShortcutsHelp: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    isOpen,
    onDismiss,
}) => {
    const [characterKeyShortcutsEnabled, setCharacterKeyShortcutsEnabled] = useTemporarySetting(
        'characterKeyShortcuts.enabled',
        true
    )
    const [enableBlobPageSwitchAreasShortcuts] = useFeatureFlag('blob-page-switch-areas-shortcuts')
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
                    {Object.values({
                        ...(enableBlobPageSwitchAreasShortcuts
                            ? KEYBOARD_SHORTCUTS
                            : omit(KEYBOARD_SHORTCUTS, Object.keys(EXPERIMENTAL_BLOB_PAGE_SHORTCUTS))),
                        ...LEGACY_KEYBOARD_SHORTCUTS,
                    })
                        .filter(({ hideInHelp }) => !hideInHelp)
                        .map(({ title, keybindings }, index) => (
                            <li
                                key={index}
                                className="list-group-item d-flex align-items-center justify-content-between"
                            >
                                {title}
                                <span>
                                    <Keybindings keybindings={keybindings} />
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

interface KeybindingProps {
    keybindings: Keybinding[]
    uppercaseOrdered?: boolean
}
export function plaintextKeybindings(keybindings: Keybinding[]): string {
    return keybindings
        .map<string>(keybinding => {
            const ordered = keybinding.ordered.map(key => key.toUpperCase())
            const joinString = isMacPlatform() ? '' : '+'
            return [...(keybinding.held || []), ...ordered].map(key => shortcutDisplayName(key)).join(joinString)
        })
        .join(' or ')
}
export const Keybindings: React.FunctionComponent<KeybindingProps> = ({ keybindings, uppercaseOrdered }) => (
    <>
        {keybindings.map((keybinding, index) => {
            const ordered = uppercaseOrdered ? keybinding.ordered.map(key => key.toUpperCase()) : keybinding.ordered
            const joinString = isMacPlatform() ? '' : '+'
            return (
                <span key={index}>
                    {index !== 0 && ' or '}
                    {[...(keybinding.held || []), ...ordered]
                        .map<React.ReactNode>((key, index) => <kbd key={index}>{shortcutDisplayName(key)}</kbd>)
                        .reduce((previous, current) => [previous, joinString, current])}
                </span>
            )
        })}
    </>
)
