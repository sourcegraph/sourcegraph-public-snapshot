import React from 'react'

import { COMMAND_PALETTE_SHORTCUTS } from '../constants'

import { NavigableList } from './NavigableList'

export const CommandPaletteModesResult: React.FC<{ onSelect: () => void }> = ({ onSelect }) => (
    <NavigableList items={COMMAND_PALETTE_SHORTCUTS}>
        {({ title, keybindings, onMatch }, { active }) => (
            <NavigableList.Item
                active={active}
                label={title}
                onClick={() => {
                    onMatch()
                    onSelect()
                }}
                keybindings={keybindings}
            />
        )}
    </NavigableList>
)
