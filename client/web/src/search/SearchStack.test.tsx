import { cleanup, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { renderWithBrandedContext, RenderWithBrandedContextResult } from '@sourcegraph/shared/src/testing'

import { useExperimentalFeatures, useSearchStackState } from '../stores'
import { SearchStackEntry } from '../stores/searchStack'

import { SearchStack } from './SearchStack'

describe('Search Stack', () => {
    const renderSearchStack = (props?: Partial<{ initialOpen: boolean }>): RenderWithBrandedContextResult =>
        renderWithBrandedContext(<SearchStack {...props} />)

    function open() {
        userEvent.click(screen.getByRole('button', { name: 'Open search session' }))
    }

    afterEach(cleanup)

    const mockEntries: SearchStackEntry[] = [
        { id: 0, type: 'search', query: 'TODO', caseSensitive: false, patternType: SearchPatternType.literal },
        { id: 1, type: 'file', path: 'path/to/file', repo: 'test', revision: 'master', lineRange: null },
    ]

    describe('inital state', () => {
        it('does not render anything if feature is disabled', () => {
            useExperimentalFeatures.setState({ enableSearchStack: false })
            useSearchStackState.setState({ addableEntry: mockEntries[0] })

            renderSearchStack()

            expect(screen.queryByRole('button', { name: 'Add search' })).not.toBeInTheDocument()
        })

        it('shows the add button if an entry can be added', () => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
            useSearchStackState.setState({ canRestoreSession: true, addableEntry: mockEntries[0] })

            expect(renderSearchStack().asFragment()).toMatchSnapshot()
        })

        it('shows the top of the stack if entries exist', () => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
            useSearchStackState.setState({ canRestoreSession: true, entries: mockEntries })

            expect(renderSearchStack().asFragment()).toMatchSnapshot()
        })
    })

    describe('restore previous session', () => {
        beforeEach(() => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
        })

        it('restores the previous session', () => {
            useSearchStackState.setState({
                entries: [],
                previousEntries: mockEntries,
                canRestoreSession: true,
                addableEntry: mockEntries[0],
            })
            renderSearchStack()
            userEvent.click(screen.getByRole('button', { name: 'Open search session' }))

            userEvent.click(screen.getByRole('button', { name: 'Restore previous session' }))
            expect(useSearchStackState.getState().entries).toEqual(mockEntries)
        })
    })

    describe('with entries', () => {
        beforeEach(() => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
            useSearchStackState.setState({
                entries: [
                    {
                        id: 0,
                        type: 'search',
                        query: 'TODO',
                        caseSensitive: false,
                        patternType: SearchPatternType.literal,
                    },
                    { id: 1, type: 'file', path: 'path/to/file', repo: 'test', revision: 'master', lineRange: null },
                ],
            })
        })

        it('opens and closes', () => {
            renderSearchStack()

            userEvent.click(screen.getByRole('button', { name: 'Open search session' }))

            const closeButtons = screen.queryAllByRole('button', { name: 'Close search session' })
            expect(closeButtons).toHaveLength(2)

            userEvent.click(closeButtons[0])
            expect(screen.queryByRole('button', { name: 'Open search session' })).toBeInTheDocument()
        })

        it('redirects to entries', () => {
            renderSearchStack()
            open()

            const entryLinks = screen.queryAllByRole('link')

            // Entries are in reverse order
            expect(entryLinks[0]).toHaveAttribute('href', '/test@master/-/blob/path/to/file')
            expect(entryLinks[1]).toHaveAttribute('href', '/search?q=TODO&patternType=literal')
        })

        it('creates notebooks', () => {
            const result = renderSearchStack()
            open()

            userEvent.click(screen.getByRole('button', { name: 'Create Notebook' }))

            expect(result.history.location.pathname).toMatchInlineSnapshot('"/notebooks/new"')
            expect(result.history.location.hash).toMatchInlineSnapshot(
                '"#query:TODO,file:http%3A%2F%2Flocalhost%2Ftest%40master%2F-%2Fblob%2Fpath%2Fto%2Ffile"'
            )
        })

        it('allows to delete entries', () => {
            renderSearchStack()
            open()

            userEvent.click(screen.getAllByRole('button', { name: 'Remove entry' })[0])
            const entryLinks = screen.queryByRole('link')
            expect(entryLinks).toBeInTheDocument()
        })

        it('opens the text annotation aria', () => {
            renderSearchStack()
            open()

            userEvent.click(screen.getAllByRole('button', { name: 'Add annotation' })[0])
            expect(screen.queryByPlaceholderText('Type to add annotation...')).toBeInTheDocument()
        })
    })

    describe('selection', () => {
        beforeEach(() => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
            useSearchStackState.setState({
                entries: [
                    {
                        id: 0,
                        type: 'search',
                        query: 'TODO',
                        caseSensitive: false,
                        patternType: SearchPatternType.literal,
                    },
                    { id: 1, type: 'file', path: 'path/to/file', repo: 'test', revision: 'master', lineRange: null },
                    {
                        id: 2,
                        type: 'search',
                        query: 'another query',
                        caseSensitive: true,
                        patternType: SearchPatternType.literal,
                    },
                    {
                        id: 3,
                        type: 'search',
                        query: 'yet another query',
                        caseSensitive: true,
                        patternType: SearchPatternType.literal,
                    },
                ],
            })
        })

        it('selects an item on click or space', () => {
            renderSearchStack()
            open()

            const item = screen.getAllByRole('option')
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
            expect(screen.queryAllByRole('option', { selected: true })).toEqual([item[1]])

            item[2].focus()
            // item1
            // item2
            // item3 <-
            // item4
            userEvent.keyboard('{space}')

            expect(screen.getByRole('option', { selected: true })).toBe(item[2])
        })

        it('selects multiple items on ctrl/meta+click/space', () => {
            renderSearchStack()
            open()

            const item = screen.getAllByRole('option')
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
            expect(screen.queryAllByRole('option', { selected: true })).toEqual([item[0], item[1]])

            item[3].focus()
            // item1 <-
            // item2 <-
            // item3
            // item4 <-
            userEvent.keyboard('{ctrl}{space}')

            expect(screen.queryAllByRole('option', { selected: true })).toEqual([item[0], item[1], item[3]])
        })

        it('selects a range of items on shift+click', () => {
            renderSearchStack()
            open()

            const item = screen.getAllByRole('option')
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

            expect(screen.queryAllByRole('option', { selected: true })).toEqual([item[0], item[1], item[2]])
        })

        it('extends the range of items on shift+click', () => {
            renderSearchStack()
            open()

            const item = screen.getAllByRole('option')
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

            expect(screen.queryAllByRole('option', { selected: true })).toEqual(item)
        })

        it('selects a range of items on shift+space', () => {
            renderSearchStack()
            open()

            const item = screen.getAllByRole('option')
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

            expect(screen.queryAllByRole('option', { selected: true })).toEqual([item[0], item[1], item[2]])
        })

        it('extends the range of items on shift+space', () => {
            renderSearchStack()
            open()

            const item = screen.getAllByRole('option')
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

            expect(screen.queryAllByRole('option', { selected: true })).toEqual(item)
        })

        it('deletes all selected entries', () => {
            renderSearchStack()
            open()

            const item = screen.getAllByRole('option')
            userEvent.click(item[0])
            userEvent.click(item[2], { shiftKey: true })
            userEvent.click(screen.queryAllByRole('button', { name: 'Remove all selected entries' })[0])

            expect(screen.queryAllByRole('option').length).toBe(1)
        })

        it('does not select entry on toggle annotion click', () => {
            renderSearchStack()
            open()

            userEvent.click(screen.queryAllByRole('button', { name: 'Add annotation' })[0])

            expect(screen.queryByRole('option', { selected: true })).not.toBeInTheDocument()
        })

        it('does not select entry on typing space into the annotation area', () => {
            renderSearchStack()
            open()

            userEvent.click(screen.queryAllByRole('button', { name: 'Add annotation' })[0])
            userEvent.type(screen.getByPlaceholderText('Type to add annotation...'), '{space}')

            expect(screen.queryByRole('option', { selected: true })).not.toBeInTheDocument()
        })
    })
})
