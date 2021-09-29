import React from 'react'

import { COMMAND_PALETTE_SHORTCUTS } from '../constants'

import { CommandPaletteResultList } from './CommandPaletteResultList'

export const CommandsModesList: React.FC<{ onSelect: () => void }> = ({ onSelect }) => (
    <CommandPaletteResultList items={COMMAND_PALETTE_SHORTCUTS}>
        {({ title, keybindings, onMatch }, { active }) => (
            <CommandPaletteResultList.Item
                active={active}
                label={title}
                onClick={() => {
                    onMatch()
                    onSelect()
                }}
                keybindings={keybindings}
            />
        )}
    </CommandPaletteResultList>
)
