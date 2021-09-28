import create from 'zustand'

import { CommandPaletteMode } from './constants'

export interface CommandPaletteState {
    isOpen: boolean
    toggleIsOpen: (options?: { open?: boolean; mode?: CommandPaletteMode }) => void
    value: string
    setValue: (value: string) => void
}

export const useCommandPaletteStore = create<CommandPaletteState>(set => ({
    isOpen: false,
    value: '',
    setValue: (value: string): void => set({ value }),
    toggleIsOpen: ({ open, mode }: { open?: boolean; mode?: CommandPaletteMode } = {}) =>
        set(({ isOpen, value }) => {
            const prefix = value[0]
            let newValue = value
            if (mode !== undefined) {
                newValue = `${mode}${prefix in CommandPaletteMode ? value.slice(1) : value}`
            }
            return { isOpen: typeof open === 'boolean' ? open : !isOpen, value: newValue }
        }),
}))
