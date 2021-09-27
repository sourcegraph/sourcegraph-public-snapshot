import { Shortcut, ShortcutProvider } from '@slimsag/react-shortcuts'
import React from 'react'

import { ActionItemAction } from '../../../actions/ActionItem'

export const ShortcutController: React.FC<{
    actions: ActionItemAction[]
    onMatch: (action: ActionItemAction) => void
}> = React.memo(({ actions, onMatch }) => (
    <ShortcutProvider>
        {actions.map((actionItem, index) => (
            <Shortcut key={index} {...actionItem.keybinding!} onMatch={() => onMatch(actionItem)} />
        ))}
    </ShortcutProvider>
))
