import { CommandItem } from './components/CommandResult'
import { useCommandPaletteStore } from './store'

export enum CommandPaletteMode {
    Fuzzy = '$',
    Command = '>',
    JumpToLine = ':',
    JumpToSymbol = '@',
    RecentSearches = '#',
}

export const COMMAND_PALETTE_COMMANDS: CommandItem[] = [
    {
        id: 'openCommandPallette',
        title: '[Beta] Command palette',
        keybindings: [{ held: ['Control'], ordered: ['k'] }],
        onClick: (): void => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true })
        },
    },
    {
        id: 'openCommandPalletteCommandMode',
        title: '[Beta] Command palette : Command mode',
        keybindings: [{ held: ['Control'], ordered: ['>'] }],
        onClick: (): void => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.Command })
        },
    },
    {
        id: 'openCommandPalletteRecentSearchesMode',
        title: '[Beta] Command palette : Recent searches mode',
        keybindings: [{ held: ['Control'], ordered: ['#'] }],
        onClick: (): void => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.RecentSearches })
        },
    },
    {
        id: 'openCommandPalletteFuzzyMode',
        title: '[Beta] Command palette : Fuzzy mode',
        keybindings: [{ held: ['Control'], ordered: ['$'] }],
        onClick: (): void => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.Fuzzy })
        },
    },
    {
        id: 'openCommandPalletteJumpToLine',
        title: '[Beta] Command palette : Jump to line mode',
        keybindings: [{ held: ['Control'], ordered: [':'] }],
        onClick: (): void => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.JumpToLine })
        },
    },
    {
        id: 'openCommandPalletteJumpToSymbol',
        title: '[Beta] Command palette : Jump to symbol mode',
        keybindings: [{ held: ['Control'], ordered: ['@'] }],
        onClick: (): void => {
            useCommandPaletteStore.getState().toggleIsOpen({ open: true, mode: CommandPaletteMode.JumpToSymbol })
        },
    },
]
