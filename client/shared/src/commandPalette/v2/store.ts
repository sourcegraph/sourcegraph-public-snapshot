import create from 'zustand'

import { CommandPaletteMode } from './constants'

interface ToggleOptions {
    open?: boolean
    mode?: CommandPaletteMode
}

export interface CommandPaletteState {
    isOpen: boolean
    toggleIsOpen: (options?: ToggleOptions) => void
    value: string
    setValue: (value: string) => void
}

export const useCommandPaletteStore = create<CommandPaletteState>(set => ({
    isOpen: false,
    value: '',
    setValue: (value: string): void => set({ value }),
    toggleIsOpen: ({ open, mode }: ToggleOptions = {}) =>
        set(({ isOpen, value }) => ({
            isOpen: typeof open === 'boolean' ? open : !isOpen,
            value: mode === value[0] ? value : mode ?? '',
            // TODO: when command palette closes, the mode selector flashes.
            // this may be this culprit.
        })),
}))
