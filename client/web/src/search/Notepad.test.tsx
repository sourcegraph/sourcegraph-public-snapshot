import { act, cleanup, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { of } from 'rxjs'
import { spy } from 'sinon'
import { vi, afterEach, beforeEach, describe, expect, it } from 'vitest'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { type RenderWithBrandedContextResult, renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import type { NotebookFields } from '../graphql-operations'
import * as backend from '../notebooks/backend'
import { useNotepadState } from '../stores'
import { addNotepadEntry, type NotepadEntry } from '../stores/notepad'

import { NotepadContainer, type NotepadProps } from './Notepad'

describe('Notepad', () => {
    const renderNotepad = (props?: Partial<NotepadProps>, enabled = true): RenderWithBrandedContextResult =>
        renderWithBrandedContext(
            <MockedTestProvider mocks={[]}>
                <MockTemporarySettings settings={{ 'search.notepad.enabled': enabled }}>
                    <NotepadContainer userId="testID" {...props} />
                </MockTemporarySettings>
            </MockedTestProvider>
        )

    function open() {
        userEvent.click(screen.getByRole('button', { name: /Open Notepad/ }))
    }

    // Notepad is hidden by default on small screens. jest-dom has no concept of
    // screen size, so this needs to be explicitly overriden to ensure that the
    // component renders.
    window.matchMedia = spy(
        (media: string): MediaQueryList => ({
            matches: true,
            media,
            addListener: () => {},
            removeListener: () => {},
            addEventListener: () => {},
            removeEventListener: () => {},
            dispatchEvent: () => false,
            onchange: null,
        })
    )

    afterEach(cleanup)

    const mockEntries: NotepadEntry[] = [
        { id: 0, type: 'search', query: 'TODO', caseSensitive: false, patternType: SearchPatternType.standard },
        {
            id: 1,
            type: 'file',
            path: 'path/to/file',
            repo: 'test',
            revision: 'master',
            lineRange: null,
            annotation: '',
        },
    ]

    describe('closed state', () => {
        it('does not render anything if feature is disabled dand there are no notes', () => {
            renderNotepad({}, false)

            expect(screen.queryByRole('complementary')).not.toBeInTheDocument()
        })

        it('shows button to open notepad', () => {
            useNotepadState.setState({ entries: mockEntries })

            renderNotepad()

            expect(screen.queryByRole('button', { name: /Open Notepad/ })).toBeInTheDocument()
        })
    })

    describe('restore previous session', () => {
        it('restores the previous session', () => {
            useNotepadState.setState({
                entries: [],
                previousEntries: mockEntries,
                canRestoreSession: true,
                addableEntry: mockEntries[0],
            })
            renderNotepad()
            userEvent.click(screen.getByRole('button', { name: /Open Notepad/ }))

            userEvent.click(screen.getByRole('button', { name: 'Restore last session' }))
            expect(useNotepadState.getState().entries).toEqual(mockEntries)
        })
    })

    describe('with notes', () => {
        beforeEach(() => {
            useNotepadState.setState({
                entries: [
                    {
                        id: 0,
                        type: 'search',
                        query: 'TODO',
                        caseSensitive: false,
                        patternType: SearchPatternType.standard,
                    },
                    { id: 1, type: 'file', path: 'path/to/file', repo: 'test', revision: 'master', lineRange: null },
                ],
            })
        })

        it('opens and closes', () => {
            renderNotepad()

            userEvent.click(screen.getByRole('button', { name: /Open Notepad/ }))
            userEvent.click(screen.getByRole('button', { name: /Close Notepad/ }))

            expect(screen.queryByRole('button', { name: /Open Notepad/ })).toBeInTheDocument()
        })

        it('redirects to notes', () => {
            renderNotepad()
            open()

            const entryLinks = screen.queryAllByRole('link')

            // Entries are in reverse order
            expect(entryLinks[0]).toHaveAttribute('href', '/test@master/-/blob/path/to/file')
            expect(entryLinks[1]).toHaveAttribute('href', '/search?q=TODO&patternType=standard&sm=0')
        })

        it('creates notebooks', () => {
            const createNotebookSpy = jest
                .spyOn(backend, 'createNotebook')
                .mockImplementation(() => of({} as unknown as NotebookFields))

            renderNotepad()
            open()

            act(() => {
                userEvent.click(screen.getByRole('button', { name: 'Create Notebook' }))
            })

            expect(createNotebookSpy).toBeCalledTimes(1)
        })

        it('allows to delete entries', () => {
            renderNotepad()
            open()

            userEvent.click(screen.getAllByRole('button', { name: 'Remove entry' })[0])
            const entryLinks = screen.queryByRole('link')
            expect(entryLinks).toBeInTheDocument()
        })

        it('opens the text annotation input', () => {
            renderNotepad()
            open()

            userEvent.click(screen.getAllByRole('button', { name: 'Add annotation' })[0])
            expect(screen.queryByPlaceholderText('Type to add annotation...')).toBeInTheDocument()
        })

        it('closes annotation input on Meta+Enter', () => {
            renderNotepad()
            open()

            userEvent.click(screen.getAllByRole('button', { name: 'Add annotation' })[0])
            userEvent.type(screen.getByPlaceholderText('Type to add annotation...'), 'test')
            userEvent.keyboard('{ctrl}{enter}')

            expect(screen.queryByPlaceholderText('Type to add annotation...')).not.toBeInTheDocument()
        })
    })

    describe('selection', () => {
        beforeEach(() => {
            useNotepadState.setState({
                entries: [
                    {
                        id: 1,
                        type: 'search',
                        query: 'TODO',
                        caseSensitive: false,
                        patternType: SearchPatternType.standard,
                    },
                    { id: 2, type: 'file', path: 'path/to/file', repo: 'test', revision: 'master', lineRange: null },
                    {
                        id: 3,
                        type: 'search',
                        query: 'another query',
                        caseSensitive: true,
                        patternType: SearchPatternType.standard,
                    },
                    {
                        id: 4,
                        type: 'search',
                        query: 'yet another query',
                        caseSensitive: true,
                        patternType: SearchPatternType.standard,
                    },
                ],
            })
        })

        it('selects an item on click or space', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            // item1 <-
            // item2
            // item3
            // item4
            userEvent.click(item[0])
            // item1
            // item2 <-
            // item3
            // item4
            userEvent.click(item[1])
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([item[1]])

            item[2].focus()
            // item1
            // item2
            // item3 <-
            // item4
            userEvent.keyboard('{space}')

            expect(screen.getByRole('listitem', { name: /^Selected/ })).toBe(item[2])
        })

        it('selects multiple items on ctrl/meta+click/space', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            // item1 <-
            // item2
            // item3
            // item4
            userEvent.click(item[0])
            // item1 <-
            // item2 <-
            // item3
            // item4
            userEvent.click(item[1], { ctrlKey: true })
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([item[0], item[1]])

            item[3].focus()
            // item1 <-
            // item2 <-
            // item3
            // item4 <-
            userEvent.keyboard('{ctrl}{space}')

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([item[0], item[1], item[3]])
        })

        it('selects a range of items on shift+click', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            // item1 <-
            // item2
            // item3
            // item4
            userEvent.click(item[0])
            // item1 <- (last)
            // item2 <-
            // item3 <-
            // item4
            userEvent.click(item[2], { shiftKey: true })

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([item[0], item[1], item[2]])
        })

        it('extends the range of items on shift+click', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            // item1
            // item2 <-
            // item3
            // item4
            userEvent.click(item[1])
            // item1
            // item2 <- (last)
            // item3 <-
            // item4 <-
            userEvent.click(item[3], { shiftKey: true })
            // Shift click always adds!
            // item1 <-
            // item2 <-
            // item3 <-
            // item4 <- (last)
            userEvent.click(item[0], { shiftKey: true })

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual(item)
        })

        it('selects a range of items on shift+space', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            // item1 <-
            // item2
            // item3
            // item4
            item[0].focus()
            userEvent.keyboard('{space}')
            // item1 <- (last)
            // item2 <-
            // item3 <-
            // item4
            item[2].focus()
            userEvent.keyboard('{shift}{space}')

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([item[0], item[1], item[2]])
        })

        it('extends the range of items on shift+space', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            // item1
            // item2 <-
            // item3
            // item4
            item[1].focus()
            userEvent.keyboard('{space}')
            // item1
            // item2 <- (last)
            // item3 <-
            // item4 <-
            item[3].focus()
            userEvent.keyboard('{shift}{space}')
            // Shift click always adds!
            // item1 <-
            // item2 <-
            // item3 <-
            // item4 <- (last)
            item[0].focus()
            userEvent.keyboard('{shift}{space}')

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual(item)
        })

        it('selects all items on ctrl+a', () => {
            renderNotepad()
            open()

            const list = screen.getByRole('list')
            const items = screen.getAllByRole('listitem')

            list.focus()
            userEvent.keyboard('{ctrl}{a}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual(items)
        })

        it('selects the next item on arrow-down', () => {
            renderNotepad()
            open()

            const list = screen.getByRole('list')
            const items = screen.getAllByRole('listitem')

            list.focus()
            userEvent.keyboard('{arrowdown}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[0]])
            expect(items[0]).toHaveFocus()

            userEvent.keyboard('{arrowdown}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[1]])
            expect(items[1]).toHaveFocus()
        })

        it('selects the previous item on arrow-up', () => {
            renderNotepad()
            open()

            const list = screen.getByRole('list')
            const items = screen.getAllByRole('listitem')

            list.focus()
            userEvent.keyboard('{arrowup}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[3]])
            expect(items[3]).toHaveFocus()

            userEvent.keyboard('{arrowup}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[2]])
            expect(items[2]).toHaveFocus()
        })

        it('extends/shrinks selection on shift+arrow-down/up', () => {
            renderNotepad()
            open()

            const list = screen.getByRole('list')
            const items = screen.getAllByRole('listitem')

            list.focus()
            userEvent.keyboard('{arrowdown}')
            userEvent.keyboard('{shift}{arrowdown}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[0], items[1]])

            userEvent.keyboard('{shift}{arrowup}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[0]])
        })

        it('skips over selected notes using shift+arrow-down', () => {
            renderNotepad()
            open()

            const items = screen.getAllByRole('listitem')

            userEvent.click(items[2], { ctrlKey: true }) // select 3. item
            userEvent.click(items[0], { ctrlKey: true }) // select 1. item

            userEvent.keyboard('{shift}{arrowdown}') // selects 2. item
            userEvent.keyboard('{shift}{arrowdown}') // selects 4. item

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([
                items[0],
                items[1],
                items[2],
                items[3],
            ])
        })

        it('skips over selected notes using shift+arrow-up', () => {
            renderNotepad()
            open()

            const items = screen.getAllByRole('listitem')

            userEvent.click(items[1], { ctrlKey: true }) // select 2. item
            userEvent.click(items[3], { ctrlKey: true }) // select 4. item

            userEvent.keyboard('{shift}{arrowdown}') // selects 3. item
            userEvent.keyboard('{shift}{arrowdown}') // selects 1. item

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([
                items[0],
                items[1],
                items[2],
                items[3],
            ])
        })

        it('extends/shrinks selection on shift+arrow-up/down', () => {
            renderNotepad()
            open()

            const list = screen.getByRole('list')
            const items = screen.getAllByRole('listitem')

            list.focus()
            userEvent.keyboard('{arrowup}')
            userEvent.keyboard('{shift}{arrowup}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[2], items[3]])

            userEvent.keyboard('{shift}{arrowdown}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[3]])
        })

        it('maintains the right selected items when non-selected items get removed', () => {
            renderNotepad()
            open()

            const items = screen.getAllByRole('listitem')
            userEvent.click(items[1])
            userEvent.click(screen.getAllByTitle('Remove entry')[0])

            // Verifies that the item is still the selected one (if not it would
            // item[2] which is now the second item).
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[1]])
        })

        it('selectes the newly added item', () => {
            renderNotepad()
            open()

            let items = screen.getAllByRole('listitem')

            // Selected 2. item
            userEvent.click(items[1])

            act(() => {
                addNotepadEntry({
                    type: 'search',
                    patternType: SearchPatternType.standard,
                    query: 'new TODO',
                    caseSensitive: false,
                })
            })

            // Referesh items
            items = screen.getAllByRole('listitem')

            expect(screen.queryAllByRole('listitem', { name: /^Selected/ })).toEqual([items[0]])
        })

        it('deletes a single entry even if others are selected', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            userEvent.click(item[0])
            userEvent.click(item[2], { shiftKey: true })
            userEvent.click(screen.queryAllByRole('button', { name: 'Remove entry' })[0])

            expect(screen.queryAllByRole('listitem').length).toBe(3)
        })

        it('deletes all selected notes when Delete is pressed', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            userEvent.click(item[0])
            userEvent.click(item[2], { shiftKey: true })
            userEvent.keyboard('{delete}')

            expect(screen.queryAllByRole('listitem').length).toBe(1)
        })

        it('clears selection on ESC', () => {
            renderNotepad()
            open()

            const item = screen.getAllByRole('listitem')
            userEvent.click(item[0])
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ }).length).toBe(1)

            userEvent.keyboard('{escape}')
            expect(screen.queryAllByRole('listitem', { name: /^Selected/ }).length).toBe(0)
        })

        it('does not select entry on toggle annotion click', () => {
            renderNotepad()
            open()

            userEvent.click(screen.queryAllByRole('button', { name: 'Add annotation' })[0])

            expect(screen.queryByRole('listitem', { name: /^Selected/ })).not.toBeInTheDocument()
        })

        it('does not select note on typing space into the annotation area', () => {
            renderNotepad()
            open()

            userEvent.click(screen.queryAllByRole('button', { name: 'Add annotation' })[0])
            userEvent.type(screen.getByPlaceholderText('Type to add annotation...'), '{space}')

            expect(screen.queryByRole('listitem', { name: /^Selected/ })).not.toBeInTheDocument()
        })
    })
})
