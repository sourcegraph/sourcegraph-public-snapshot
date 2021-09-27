import create from 'zustand'

export interface CommandPaletteState {
    isOpen: boolean
    toggleIsOpen: () => void
}

export const useCommandPaletteStore = create<CommandPaletteState>(set => ({
    isOpen: false,
    toggleIsOpen: () => set(state => ({ isOpen: !state.isOpen })),
}))
