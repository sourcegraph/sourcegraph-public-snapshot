import { KeyboardShortcut } from '../../../keyboardShortcuts'
import { CommandPaletteMode } from '../constants'

import { useCommandPaletteStore } from '../store'

export type KeyboardShortcutWithCallback = KeyboardShortcut & { onMatch: () => void }

export const COMMAND_PALETTE_SHORTCUTS: KeyboardShortcutWithCallback[] = [
    {
        id: 'openCommandPallette',
        title: 'Command palette',
        keybindings: [{ held: ['Control'], ordered: ['k'] }],
        onMatch: () => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true })
        },
    },
    {
        id: 'openCommandPalletteCommandMode',
        title: 'Command palette -> command mode',
        keybindings: [{ held: ['Control'], ordered: ['>'] }],
        onMatch: () => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.Command })
        },
    },
    {
        id: 'openCommandPalletteRecentSearchesMode',
        title: 'Command palette -> recent searches mode',
        keybindings: [{ held: ['Control'], ordered: ['#'] }],
        onMatch: () => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.RecentSearches })
        },
    },
    {
        id: 'openCommandPalletteFuzzyMode',
        title: 'Command palette -> fuzzy mode',
        keybindings: [{ held: ['Control'], ordered: ['$'] }],
        onMatch: () => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.Fuzzy })
        },
    },
    {
        id: 'openCommandPalletteJumpToLine',
        title: 'Command palette -> jump to line mode',
        keybindings: [{ held: ['Control'], ordered: [':'] }],
        onMatch: () => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.JumpToLine })
        },
    },
    {
        id: 'openCommandPalletteJumpToSymbol',
        title: 'Command palette -> jump to symbol mode',
        keybindings: [{ held: ['Control'], ordered: ['@'] }],
        onMatch: () => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.Fuzzy })
        },
    },
]
