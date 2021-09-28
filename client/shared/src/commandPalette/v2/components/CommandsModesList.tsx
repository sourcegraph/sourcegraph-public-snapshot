import React from 'react'

import { COMMAND_PALETTE_SHORTCUTS } from '../constants'

import { CommandPaletteResultList } from './CommandPaletteResultList'

export const CommandsModesList: React.FC = () => (
    <CommandPaletteResultList>
        {COMMAND_PALETTE_SHORTCUTS.map(({ id, title, keybindings, onMatch }) => (
            <CommandPaletteResultList.Item key={id} label={title} onClick={onMatch} keybindings={keybindings} />
        ))}
    </CommandPaletteResultList>
)
