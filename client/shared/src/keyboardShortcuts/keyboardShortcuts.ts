import { isMacPlatform } from '@sourcegraph/common'

import { KeyboardShortcut } from '../keyboardShortcuts'

type KEYBOARD_SHORTCUT_IDENTIFIERS =
    | 'commandPalette'
    | 'switchTheme'
    | 'keyboardShortcutsHelp'
    | 'focusSearch'
    | 'fuzzyFinder'
    | 'copyFullQuery'

export type KEYBOARD_SHORTCUT_MAPPING = Record<KEYBOARD_SHORTCUT_IDENTIFIERS, KeyboardShortcut>

export const KEYBOARD_SHORTCUTS: KEYBOARD_SHORTCUT_MAPPING = {
    commandPalette: {
        title: 'Show command palette',
        keybindings: [{ held: ['Control'], ordered: ['p'] }, { ordered: ['F1'] }, { held: ['Alt'], ordered: ['x'] }],
    },
    switchTheme: {
        title: 'Switch color theme',
        // use '†' here to make `Alt + t` works on macos
        keybindings: [{ held: ['Alt'], ordered: [isMacPlatform() ? ('†' as any) : 't'] }],
    },
    keyboardShortcutsHelp: {
        title: 'Show keyboard shortcuts help',
        keybindings: [{ held: ['Shift'], ordered: ['?'] }],
        hideInHelp: true,
    },
    focusSearch: {
        title: 'Focus search bar',
        keybindings: [{ ordered: ['/'] }],
    },
    fuzzyFinder: {
        title: 'Fuzzy search files',
        keybindings: [{ held: [isMacPlatform() ? 'Meta' : 'Control'], ordered: ['k'] }],
    },
    copyFullQuery: {
        title: 'Copy full query',
        keybindings: [{ held: [isMacPlatform() ? 'Meta' : 'Control', 'Shift'], ordered: ['c'] }],
    },
}
