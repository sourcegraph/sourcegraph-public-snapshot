import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { act } from 'react-dom/test-utils'
import sinon from 'sinon'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/testing/MockIntersectionObserver'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

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
        platformContext: NOOP_PLATFORM_CONTEXT,
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
        const props = { ...defaultProps, isExternalServicesUserModeAll: true }

        it('should not display CTA if not on Sourcegraph.com', () => {
            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextDropdown {...props} isSourcegraphDotCom={false} hasUserAddedRepositories={false} />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context/ }))

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should display CTA on Sourcegraph.com if no repos have been added and not permanently dismissed', () => {
            renderWithBrandedContext(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextDropdown {...props} isSourcegraphDotCom={true} hasUserAddedRepositories={false} />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context/ }))

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
                        {...props}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                        authenticatedUser={mockUserWithOrg}
                    />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context/ }))

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if repos have been added', () => {
            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextDropdown {...props} isSourcegraphDotCom={true} hasUserAddedRepositories={true} />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context/ }))

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if dimissed', () => {
            renderWithBrandedContext(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': true }}>
                    <SearchContextDropdown {...props} isSourcegraphDotCom={true} hasUserAddedRepositories={false} />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context/ }))

            expect(screen.queryByRole('button', { name: /Don't show this againr/ })).not.toBeInTheDocument()
        })

        it('should dismiss CTA when clicking dismiss button', async () => {
            const onSettingsChanged = sinon.spy()

            renderWithBrandedContext(
                <MockTemporarySettings
                    settings={{ 'search.contexts.ctaDismissed': false }}
                    onSettingsChanged={onSettingsChanged}
                >
                    <SearchContextDropdown {...props} isSourcegraphDotCom={true} hasUserAddedRepositories={false} />
                </MockTemporarySettings>
            )

            userEvent.click(screen.getByRole('button', { name: /context/ }))

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
