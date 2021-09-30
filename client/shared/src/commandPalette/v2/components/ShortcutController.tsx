import { Shortcut, ShortcutProvider } from '@slimsag/react-shortcuts'
import React from 'react'

import { KeyboardShortcut } from '../../../keyboardShortcuts'

export type KeyboardShortcutWithCallback = KeyboardShortcut & { onMatch: () => void }

export const ShortcutController: React.FC<{
    shortcuts: KeyboardShortcutWithCallback[]
}> = React.memo(({ shortcuts }) => (
    <ShortcutProvider>
        {shortcuts.map(({ keybindings, onMatch, id }) =>
            keybindings.map((keybinding, index) => (
                <Shortcut key={`${id}-${index}`} {...keybinding} onMatch={onMatch} />
            ))
        )}
    </ShortcutProvider>
))
