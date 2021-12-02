import { ShortcutProvider } from '@slimsag/react-shortcuts'
import { render, fireEvent, waitFor, screen } from '@testing-library/react'
import React from 'react'

import { KeyboardShortcutsHelp } from './KeyboardShortcutsHelp'

describe('KeyboardShortcutsHelp', () => {
    test('is triggered correctly', async () => {
        render(
            <ShortcutProvider>
                <KeyboardShortcutsHelp
                    keyboardShortcuts={[
                        {
                            id: 'x',
                            title: 't',
                            keybindings: [{ ordered: ['x'] }],
                        },
                    ]}
                    keyboardShortcutForShow={{
                        id: 'x',
                        title: 't',
                        keybindings: [{ ordered: ['x'] }],
                    }}
                />
            </ShortcutProvider>
        )

        // couldn't trigger event with ctrl/alt/shift key so use shortcut without held keys
        fireEvent.keyDown(document, { key: 'x', keyCode: 88 })

        await waitFor(() => {
            expect(screen.getByText(/keyboard shortcuts/i)).toBeInTheDocument()
            expect(screen.getByRole('dialog')).toHaveClass('show')
        })

        expect(document.body).toMatchSnapshot()
    })
})
