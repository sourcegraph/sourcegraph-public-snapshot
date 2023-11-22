import { render } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { MockTemporarySettings } from '../settings/temporary/testUtils'

import type { KEYBOARD_SHORTCUTS } from './keyboardShortcuts'
import { useKeyboardShortcut } from './useKeyboardShortcut'

const ShortcutUsageExample = ({ shortcut }: { shortcut: keyof typeof KEYBOARD_SHORTCUTS }) => {
    const keyboardShortcut = useKeyboardShortcut(shortcut)

    if (!keyboardShortcut) {
        return <span>Keyboard shortcut not found</span>
    }

    return <pre>{JSON.stringify(keyboardShortcut, null, 2)}</pre>
}

describe('useKeyboardShortcut', () => {
    describe('when character key shortcuts are enabled', () => {
        it('should return character keyboard shortcuts', () => {
            const wrapper = render(
                <MockTemporarySettings settings={{ 'characterKeyShortcuts.enabled': true }}>
                    <ShortcutUsageExample shortcut="focusSearch" />
                </MockTemporarySettings>
            )

            expect(wrapper.container).toMatchInlineSnapshot(`
                <div>
                  <pre>
                    {
                  "title": "Focus search bar",
                  "keybindings": [
                    {
                      "ordered": [
                        "/"
                      ]
                    }
                  ]
                }
                  </pre>
                </div>
            `)
        })

        it('should still return the keyboard shortcuts that contain modifiers', () => {
            const wrapper = render(
                <MockTemporarySettings settings={{ 'characterKeyShortcuts.enabled': false }}>
                    <ShortcutUsageExample shortcut="fuzzyFinder" />
                </MockTemporarySettings>
            )

            expect(wrapper.container).toMatchInlineSnapshot(`
                <div>
                  <pre>
                    {
                  "title": "Fuzzy finder",
                  "keybindings": [
                    {
                      "held": [
                        "Mod"
                      ],
                      "ordered": [
                        "k"
                      ]
                    }
                  ]
                }
                  </pre>
                </div>
            `)
        })
    })

    describe('when character key shortcuts are disabled', () => {
        it('should NOT return the character keyboard shortcut', () => {
            const wrapper = render(
                <MockTemporarySettings settings={{ 'characterKeyShortcuts.enabled': false }}>
                    <ShortcutUsageExample shortcut="focusSearch" />
                </MockTemporarySettings>
            )

            expect(wrapper.container).toMatchInlineSnapshot(`
                <div>
                  <span>
                    Keyboard shortcut not found
                  </span>
                </div>
            `)
        })

        it('should still return the keyboard shortcuts that contain modifiers', () => {
            const wrapper = render(
                <MockTemporarySettings settings={{ 'characterKeyShortcuts.enabled': false }}>
                    <ShortcutUsageExample shortcut="fuzzyFinder" />
                </MockTemporarySettings>
            )

            expect(wrapper.container).toMatchInlineSnapshot(`
                <div>
                  <pre>
                    {
                  "title": "Fuzzy finder",
                  "keybindings": [
                    {
                      "held": [
                        "Mod"
                      ],
                      "ordered": [
                        "k"
                      ]
                    }
                  ]
                }
                  </pre>
                </div>
            `)
        })
    })
})
