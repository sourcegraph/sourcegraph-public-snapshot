import { isMacPlatform } from '@sourcegraph/common'

import { KeyboardShortcut } from '../keyboardShortcuts'

export const KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE: KeyboardShortcut = {
    id: 'commandPalette',
    title: 'Show command palette',
    keybindings: [{ held: ['Control'], ordered: ['p'] }, { ordered: ['F1'] }, { held: ['Alt'], ordered: ['x'] }],
}

export const KEYBOARD_SHORTCUT_SWITCH_THEME: KeyboardShortcut = {
    id: 'switchTheme',
    title: 'Switch color theme',
    keybindings: [{ held: ['Alt'], ordered: ['t'] }],
}

export const KEYBOARD_SHORTCUT_SHOW_HELP: KeyboardShortcut = {
    id: 'keyboardShortcutsHelp',
    title: 'Show keyboard shortcuts help',
    keybindings: [{ held: ['Shift'], ordered: ['?'] }],
    hideInHelp: true,
}

export const KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR: KeyboardShortcut = {
    id: 'focusSearch',
    title: 'Focus search bar',
    keybindings: [{ ordered: ['/'] }],
}

export const KEYBOARD_SHORTCUT_FUZZY_FINDER: KeyboardShortcut = {
    id: 'fuzzyFinder',
    title: 'Fuzzy search files',
    keybindings: [{ held: [isMacPlatform() ? 'Meta' : 'Control'], ordered: ['k'] }],
}

export const KEYBOARD_SHORTCUT_CLOSE_FUZZY_FINDER: KeyboardShortcut = {
    id: 'closeFuzzyFiles',
    title: 'Close fuzzy search files',
    keybindings: [{ ordered: ['Escape'] }],
}

export const KEYBOARD_SHORTCUT_COPY_FULL_QUERY: KeyboardShortcut = {
    id: 'copyFullQuery',
    title: 'Copy full query',
    keybindings: [{ held: [isMacPlatform() ? 'Meta' : 'Control', 'Shift'], ordered: ['c'] }],
}

/**
 * Global keyboard shortcuts. React components should access these via {@link KeybindingsProps}, not
 * globally.
 */
export const KEYBOARD_SHORTCUTS: KeyboardShortcut[] = [
    KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE,
    KEYBOARD_SHORTCUT_SWITCH_THEME,
    KEYBOARD_SHORTCUT_SHOW_HELP,
    KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR,
    KEYBOARD_SHORTCUT_FUZZY_FINDER,
    KEYBOARD_SHORTCUT_CLOSE_FUZZY_FINDER,
    KEYBOARD_SHORTCUT_COPY_FULL_QUERY,
]

/** A partial React props type for components that use or propagate keyboard shortcuts. */
export interface KeyboardShortcutsProps {
    /** The global map of keybindings and their associated actions. */
    keyboardShortcuts: KeyboardShortcut[]
}
