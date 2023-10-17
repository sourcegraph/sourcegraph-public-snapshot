import type { KeyboardShortcut } from '../keyboardShortcuts'
import { useTemporarySetting } from '../settings/temporary'

import { KEYBOARD_SHORTCUTS } from './keyboardShortcuts'

type KeyboardShortcutKey = keyof typeof KEYBOARD_SHORTCUTS

/**
 * Extract a keyboard shortcut from the given key.
 * Should be used with the <Shortcut> component.
 *
 * Note: This hook supports filtering out any character-key shortcuts.
 * This is [required for WCAG 2.1 compliance](https://www.w3.org/WAI/WCAG21/Techniques/general/G217).
 * Due to this, you cannot rely on a character-key shortcut *always* existing.
 */
export function useKeyboardShortcut(key: KeyboardShortcutKey): KeyboardShortcut | undefined {
    const [characterKeyboardShortcutsEnabled] = useTemporarySetting('characterKeyShortcuts.enabled', true)

    const shortcut = KEYBOARD_SHORTCUTS[key]

    if (!characterKeyboardShortcutsEnabled) {
        // Preserve any keybindings that include modifier keys, as they are still WCAG compliant.
        const filteredKeybindings = shortcut.keybindings.filter(({ held }) => held)

        return filteredKeybindings.length ? { ...shortcut, keybindings: filteredKeybindings } : undefined
    }

    return shortcut
}
