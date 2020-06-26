import React from 'react'
import { KeyboardShortcutsHelp } from './KeyboardShortcutsHelp'
import { mount } from 'enzyme'

describe('KeyboardShortcutsHelp', () => {
    test('', () => {
        const output = mount(
            <KeyboardShortcutsHelp
                keyboardShortcuts={[
                    {
                        id: 'x',
                        title: 't',
                        keybindings: [{ held: ['Alt'], ordered: ['x'] }],
                    },
                ]}
                keyboardShortcutForShow={{
                    id: 'x',
                    title: 't',
                    keybindings: [{ held: ['Alt'], ordered: ['x'] }],
                }}
                forceIsOpen={true}
            />
        )
        expect(output).toMatchSnapshot()
    })
})
