import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import { act } from 'react-dom/test-utils'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/util/MockIntersectionObserver'

import { AuthenticatedUser } from '../../auth'
import { SourcegraphContext } from '../../jscontext'
import { MockTemporarySettings } from '../../settings/temporary/testUtils'

import { SearchContextDropdown, SearchContextDropdownProps } from './SearchContextDropdown'

describe('SearchContextDropdown', () => {
    const defaultProps: SearchContextDropdownProps = {
        telemetryService: NOOP_TELEMETRY_SERVICE,
        query: '',
        showSearchContextManagement: false,
        fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(1),
        fetchSearchContexts: mockFetchSearchContexts,
        getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
        defaultSearchContextSpec: '',
        selectedSearchContextSpec: '',
        setSelectedSearchContextSpec: () => {},
        hasUserAddedRepositories: false,
        hasUserAddedExternalServices: false,
        isSourcegraphDotCom: false,
        authenticatedUser: null,
        searchContextsEnabled: true,
    }
    const RealIntersectionObserver = window.IntersectionObserver
    let clock: sinon.SinonFakeTimers

    beforeAll(() => {
        clock = sinon.useFakeTimers()
        window.IntersectionObserver = MockIntersectionObserver
    })

    afterAll(() => {
        clock.restore()
        window.IntersectionObserver = RealIntersectionObserver
    })

    it('should start closed', () => {
        render(<SearchContextDropdown {...defaultProps} />)
        expect(screen.getByTestId('dropdown')).not.toHaveClass('show')
    })

    it('should open when toggle event happens', () => {
        render(<SearchContextDropdown {...defaultProps} />)
        userEvent.click(screen.getByTestId('dropdown-toggle'))

        expect(screen.getByTestId('dropdown')).toHaveClass('show')
    })

    it('should close if toggle event happens again', () => {
        render(<SearchContextDropdown {...defaultProps} />)

        // Click to open
        userEvent.click(screen.getByTestId('dropdown-toggle'))

        // Click to close
        userEvent.click(screen.getByTestId('dropdown-toggle'))

        expect(screen.getByTestId('dropdown')).not.toHaveClass('show')
    })

    it('should be enabled if query is empty', () => {
        render(<SearchContextDropdown {...defaultProps} />)
        expect(screen.getByTestId('dropdown-toggle')).toBeEnabled()
        expect(screen.getByTestId('dropdown-toggle')).toHaveAttribute('data-tooltip', '')
    })

    it('should be enabled if query does not contain context filter', () => {
        render(<SearchContextDropdown {...defaultProps} query="test (repo:foo or repo:python)" />)
        expect(screen.getByTestId('dropdown-toggle')).toBeEnabled()
        expect(screen.getByTestId('dropdown-toggle')).toHaveAttribute('data-tooltip', '')
    })

    it('should be disabled if query contains context filter', () => {
        render(<SearchContextDropdown {...defaultProps} query="test (context:foo or repo:python)" />)
        expect(screen.getByTestId('dropdown-toggle')).toBeDisabled()
        expect(screen.getByTestId('dropdown-toggle')).toHaveAttribute('data-tooltip', 'Overridden by query')
    })

    it('should submit search on item click', () => {
        const submitSearch = sinon.spy()
        const { rerender } = render(
            <SearchContextDropdown {...defaultProps} submitSearch={submitSearch} query="test" />
        )

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        rerender(<SearchContextDropdown {...defaultProps} submitSearch={submitSearch} query="test" />)

        userEvent.click(screen.getByTestId('search-context-menu-item'))

        sinon.assert.calledOnce(submitSearch)
    })

    describe('with CTA', () => {
        let oldContext: SourcegraphContext & Mocha.SuiteFunction
        beforeEach(() => {
            oldContext = window.context
            window.context = { externalServicesUserMode: 'all' } as SourcegraphContext & Mocha.SuiteFunction
        })

        afterEach(() => {
            window.context = oldContext
        })

        it('should not display CTA if not on Sourcegraph.com', () => {
            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={false}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context:/ }))

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should display CTA on Sourcegraph.com if no repos have been added and not permanently dismissed', () => {
            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context:/ }))

            expect(screen.getByRole('button', { name: /Don't show this again/ })).toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if user is part of an org', () => {
            const mockUserWithOrg = {
                organizations: {
                    nodes: [{ displayName: 'test org', id: '1', name: 'test' }],
                },
            } as AuthenticatedUser

            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                        authenticatedUser={mockUserWithOrg}
                    />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context:/ }))

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if repos have been added', () => {
            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={true}
                    />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context:/ }))

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if dimissed', () => {
            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': true }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context:/ }))

            expect(screen.queryByRole('button', { name: /Don't show this againr/ })).not.toBeInTheDocument()
        })

        it('should dismiss CTA when clicking dismiss button', async () => {
            const onSettingsChanged = sinon.spy()

            render(
                <MockTemporarySettings
                    settings={{ 'search.contexts.ctaDismissed': false }}
                    onSettingsChanged={onSettingsChanged}
                >
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context:/ }))

            // would need some time for animation before the button becomes clickable
            // otherwise we would get `unable to click element as it has or inherits pointer-events set to "none".` error
            await waitFor(() =>
                userEvent.click(screen.getByRole('button', { name: /Don't show this again/ }), undefined, {
                    skipPointerEventsCheck: true,
                })
            )

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
            expect(screen.getByRole('searchbox')).toBeInTheDocument()

            sinon.assert.calledOnceWithExactly(onSettingsChanged, { 'search.contexts.ctaDismissed': true })
        })
    })
})
