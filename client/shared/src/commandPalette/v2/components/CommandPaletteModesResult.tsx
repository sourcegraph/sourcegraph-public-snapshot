import React from 'react'

import { COMMAND_PALETTE_COMMANDS } from '../constants'

import { NavigableList } from './NavigableList'

export const CommandPaletteModesResult: React.FC<{ onSelect: () => void }> = ({ onSelect }) => (
    <NavigableList items={COMMAND_PALETTE_COMMANDS}>
        {({ title, keybindings, onClick }, { active }) => (
            <NavigableList.Item
                active={active}
                onClick={() => {
                    onClick()
                    onSelect()
                }}
                keybindings={keybindings}
            >
                {title}
            </NavigableList.Item>
        )}
    </NavigableList>
)
