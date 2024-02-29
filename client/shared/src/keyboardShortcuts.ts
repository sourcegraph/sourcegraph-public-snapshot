import { isMacPlatform } from '@sourcegraph/common'

import type { ModifierKey, Key } from './react-shortcuts/keys'
import { getModKey } from './react-shortcuts/ShortcutManager'

/**
 * An action and its associated keybindings.
 */
export interface KeyboardShortcut {
    /** A descriptive title. */
    title: string

    /** The keybindings that trigger this shortcut. */
    keybindings: Keybinding[]

    /** If set, do not show this in the KeyboardShortcutsHelp modal. */
    hideInHelp?: boolean
}

/** A key sequence (that triggers a keyboard shortcut). */
export interface Keybinding {
    /** Keys that must be held down. */
    held?: (ModifierKey | 'Mod')[]

    /** Keys that must be pressed in order (when holding the `held` keys). */
    ordered: Key[]
}

const isMacOS = isMacPlatform()

const KEY_TO_NAME: { [P in Key | ModifierKey | string]?: string } = {
    Meta: isMacOS ? '⌘' : 'Cmd',
    Shift: isMacOS ? '⇧' : 'Shift',
    Control: isMacOS ? '^' : 'Ctrl',
    '†': 't',
    ArrowUp: '↑',
    ArrowDown: '↓',
    Enter: isMacOS ? 'Return' : 'Enter',
    Backspace: isMacOS ? '⌫' : 'Backspace',
}
KEY_TO_NAME.Mod = KEY_TO_NAME[getModKey()]

const keySeparator = isMacOS ? ' ' : '+'

/**
 * Returns the platform specific sequence of name/symbol for the provided key
 * binding. The input needs to be in the form of `<key>+<key>+<key>...`, where
 * <key> is the name of a key (e.g. Shift or a).
 */
export function shortcutDisplayName(sequence: string): string {
    return sequence
        .split(/\s*\+\s*/)
        .map(key => KEY_TO_NAME[key] ?? key)
        .join(keySeparator)
}
