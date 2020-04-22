import { KeyboardShortcut } from '../../../shared/src/keyboardShortcuts'

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

/**
 * Global keyboard shortcuts. React components should access these via {@link KeybindingsProps}, not
 * globally.
 */
export const KEYBOARD_SHORTCUTS: KeyboardShortcut[] = [
    KEYBOARD_SHORTCUT_SHOW_COMMAND_PALETTE,
    KEYBOARD_SHORTCUT_SWITCH_THEME,
    KEYBOARD_SHORTCUT_SHOW_HELP,
]

/** A partial React props type for components that use or propagate keyboard shortcuts. */
export interface KeyboardShortcutsProps {
    /** The global map of keybindings and their associated actions. */
    keyboardShortcuts: KeyboardShortcut[]
}
