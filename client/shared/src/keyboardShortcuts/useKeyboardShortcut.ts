import { KeyboardShortcut } from '../keyboardShortcuts'
import { useTemporarySetting } from '../settings/temporary/useTemporarySetting'

import { KEYBOARD_SHORTCUTS } from './keyboardShortcuts'

type KeyboardShortcutKey = keyof typeof KEYBOARD_SHORTCUTS

export function useKeyboardShortcut(key: KeyboardShortcutKey): KeyboardShortcut | undefined {
    const [keyboardShortcutsEnabled] = useTemporarySetting('keyboardShortcuts.enabled', true)

    if (!keyboardShortcutsEnabled) {
        return undefined
    }

    return KEYBOARD_SHORTCUTS[key]
}
