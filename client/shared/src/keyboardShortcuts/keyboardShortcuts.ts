import { isMacPlatform, isSafari } from '@sourcegraph/common'

import type { KeyboardShortcut } from '../keyboardShortcuts'

type KEYBOARD_SHORTCUT_IDENTIFIERS =
    | 'switchTheme'
    | 'keyboardShortcutsHelp'
    | 'focusSearch'
    | 'fuzzyFinder'
    | 'fuzzyFinderActions'
    | 'fuzzyFinderRepos'
    | 'fuzzyFinderSymbols'
    | 'fuzzyFinderFiles'
    | 'focusCodeEditor'
    | 'focusFileTree'
    | 'focusSymbols'
    | 'focusCody'

export type KEYBOARD_SHORTCUT_MAPPING = Record<KEYBOARD_SHORTCUT_IDENTIFIERS, KeyboardShortcut>

export const EXPERIMENTAL_BLOB_PAGE_SHORTCUTS: Record<
    'focusCodeEditor' | 'focusFileTree' | 'focusSymbols' | 'focusCody',
    KeyboardShortcut
> = {
    focusCodeEditor: {
        title: 'Focus editor',
        keybindings: [{ ordered: ['c'] }],
    },
    focusFileTree: {
        title: 'Focus file tree',
        keybindings: [{ ordered: ['f'] }],
    },
    focusSymbols: {
        title: 'Focus symbols',
        keybindings: [{ ordered: ['s'] }],
    },
    focusCody: {
        title: 'Focus Cody',
        keybindings: [{ held: ['Alt'], ordered: ['/'] }],
    },
}

export const KEYBOARD_SHORTCUTS: KEYBOARD_SHORTCUT_MAPPING = {
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
    ...EXPERIMENTAL_BLOB_PAGE_SHORTCUTS,
    fuzzyFinder: {
        title: 'Fuzzy finder',
        keybindings: [{ held: ['Mod'], ordered: ['k'] }],
    },
    fuzzyFinderActions: {
        title: 'Fuzzy find actions',
        keybindings: [{ held: ['Mod', 'Shift'], ordered: ['a'] }],
        hideInHelp: true,
    },
    fuzzyFinderRepos: {
        title: 'Fuzzy find repos',
        keybindings: [{ held: ['Mod'], ordered: ['i'] }],
        hideInHelp: true,
    },
    fuzzyFinderFiles: {
        title: 'Fuzzy find files',
        keybindings: [{ held: ['Mod'], ordered: ['p'] }],
        hideInHelp: true,
    },
    fuzzyFinderSymbols: {
        title: 'Fuzzy find symbols',
        keybindings: [{ held: isSafari() ? ['Mod', 'Shift'] : ['Mod'], ordered: ['o'] }],
        hideInHelp: true,
    },
}
