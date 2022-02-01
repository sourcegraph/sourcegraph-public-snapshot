import { cleanup, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

import { useExperimentalFeatures, useSearchStackState } from '../stores'
import { SearchStackEntry } from '../stores/searchStack'

import { SearchStack } from './SearchStack'

describe('Search Stack', () => {
    const renderSearchStack = (props?: Partial<{ initialOpen: boolean }>): RenderWithRouterResult =>
        renderWithRouter(<SearchStack {...props} />)

    afterEach(cleanup)

    const mockEntries: SearchStackEntry[] = [
        { type: 'search', query: 'TODO', caseSensitive: false, patternType: SearchPatternType.literal },
        { type: 'file', path: 'path/to/file', repo: 'test', revision: 'master', lineRange: null },
    ]

    describe('inital state', () => {
        it('does not render anything if feature is disabled', () => {
            useExperimentalFeatures.setState({ enableSearchStack: false })

            renderSearchStack()

            expect(screen.queryByRole('button', { name: 'Open search session' })).not.toBeInTheDocument()
        })

        it('does not render anything if there is no previous session', () => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
            useSearchStackState.setState({ canRestoreSession: false })

            renderSearchStack()

            expect(screen.queryByRole('button', { name: 'Open search session' })).not.toBeInTheDocument()
        })

        it('renders something if a previous session can be restored', () => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
            useSearchStackState.setState({ canRestoreSession: true })

            renderSearchStack()

            expect(screen.queryByRole('button', { name: 'Open search session' })).toBeInTheDocument()
        })
    })

    describe('restore previous session', () => {
        beforeEach(() => {
            useExperimentalFeatures.setState({ enableSearchStack: true })
        })

        it('restores the previous session', () => {
            useSearchStackState.setState({ entries: [], previousEntries: mockEntries, canRestoreSession: true })
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
                    { type: 'search', query: 'TODO', caseSensitive: false, patternType: SearchPatternType.literal },
                    { type: 'file', path: 'path/to/file', repo: 'test', revision: 'master', lineRange: null },
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
            const result = renderSearchStack()
            userEvent.click(screen.getByRole('button', { name: 'Open search session' }))

            const entryLinks = screen.queryAllByRole('link')

            userEvent.click(entryLinks[0])
            expect(result.history.location.pathname).toMatchInlineSnapshot('"/search"')
            expect(result.history.location.search).toMatchInlineSnapshot('"?q=TODO&patternType=literal"')

            userEvent.click(entryLinks[1])
            expect(result.history.location.pathname).toMatchInlineSnapshot('"/test@master/-/blob/path/to/file"')
            expect(result.history.location.search).toMatchInlineSnapshot('""')
        })

        it('creates notebooks', () => {
            const result = renderSearchStack()

            userEvent.click(screen.getByRole('button', { name: 'Open search session' }))
            userEvent.click(screen.getByRole('button', { name: 'Create Notebook' }))

            expect(result.history.location.pathname).toMatchInlineSnapshot('"/notebooks/new"')
            expect(result.history.location.hash).toMatchInlineSnapshot(
                '"#query:TODO,file:http%3A%2F%2Flocalhost%2Ftest%40master%2F-%2Fblob%2Fpath%2Fto%2Ffile"'
            )
        })
    })
})
