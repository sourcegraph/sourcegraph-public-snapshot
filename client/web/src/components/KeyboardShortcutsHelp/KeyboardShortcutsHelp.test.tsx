import { useState } from 'react'

import { fireEvent, screen } from '@testing-library/react'

import { Shortcut, ShortcutProvider } from '@sourcegraph/shared/src/react-shortcuts'
import { renderWithBrandedContext } from '@sourcegraph/wildcard'

import { KeyboardShortcutsHelp } from './KeyboardShortcutsHelp'

// We don't use the original show help shortcut as we cannot trigger shortcuts with 'Held' modifiers
const showHelpShortcut = 'Y'

const ShortcutTriggerExample = () => {
    const [isOpen, setIsOpen] = useState(false)
    return (
        <ShortcutProvider>
            <Shortcut ordered={[showHelpShortcut]} onMatch={() => setIsOpen(true)} />
            <KeyboardShortcutsHelp isOpen={isOpen} onDismiss={() => setIsOpen(false)} />
        </ShortcutProvider>
    )
}

describe('KeyboardShortcutsHelp', () => {
    test('is triggered correctly', () => {
        renderWithBrandedContext(<ShortcutTriggerExample />)

        // Enable the help modal
        fireEvent.keyDown(document.body, { key: showHelpShortcut })

        expect(screen.getByRole('heading', { name: /keyboard shortcuts/i })).toBeVisible()
        expect(document.body).toMatchSnapshot()
    })
})
