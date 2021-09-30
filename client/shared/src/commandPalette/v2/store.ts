import React, { useEffect } from 'react'
import create from 'zustand'

import { KEYBOARD_SHORTCUT_SWITCH_THEME } from '@sourcegraph/web/src/keyboardShortcuts/keyboardShortcuts'

import { CommandItem } from './components/CommandResult'
import { CommandPaletteMode } from './constants'
import { useDeepCompareEffectNoCheck } from 'use-deep-compare-effect'

interface ToggleOptions {
    open?: boolean
    mode?: CommandPaletteMode
}

export interface CommandPaletteState {
    isOpen: boolean
    toggleIsOpen: (options?: ToggleOptions) => void
    value: string
    setValue: (value: string) => void
    extraCommands: CommandItem[]
    addCommand: (command: CommandItem) => () => void
}

const BUILT_IN_COMMANDS: CommandItem[] = [
    {
        ...KEYBOARD_SHORTCUT_SWITCH_THEME,
        onClick: KEYBOARD_SHORTCUT_SWITCH_THEME.onMatch,
    },
]

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
    extraCommands: [...BUILT_IN_COMMANDS],
    addCommand: (command: CommandItem) => {
        set(({ extraCommands }) => ({
            extraCommands: [...extraCommands, command],
        }))

        return () => {
            set(({ extraCommands }) => ({
                extraCommands: extraCommands.filter(extraCommand => extraCommand !== command),
            }))
        }
    },
}))

export const BuiltInCommand: React.FC<{ commandItem: CommandItem }> = ({ commandItem }) => {
    const { addCommand } = useCommandPaletteStore()

    // TEMP
    // Make sure consumers memoize commandItem
    useDeepCompareEffectNoCheck(() => addCommand(commandItem), [commandItem, addCommand])

    return null
}
