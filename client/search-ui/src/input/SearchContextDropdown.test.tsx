import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { act } from 'react-dom/test-utils'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
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
        expect(screen.queryByTestId('dropdown-content')).not.toBeInTheDocument()
    })

    it('should open when toggle event happens', () => {
        render(<SearchContextDropdown {...defaultProps} />)
        userEvent.click(screen.getByTestId('dropdown-toggle'))

        expect(screen.queryByTestId('dropdown-content')).toBeInTheDocument()
    })

    it('should close if toggle event happens again', () => {
        render(<SearchContextDropdown {...defaultProps} />)

        // Click to open
        userEvent.click(screen.getByTestId('dropdown-toggle'))

        // Click to close
        userEvent.click(screen.getByTestId('dropdown-toggle'))

        expect(screen.queryByTestId('dropdown-content')).not.toBeInTheDocument()
    })

    it('should be enabled if query is empty', () => {
        render(<SearchContextDropdown {...defaultProps} />)
        expect(screen.getByTestId('dropdown-toggle')).toBeEnabled()
        expect(screen.getByTestId('dropdown-toggle')).toHaveAttribute('data-test-tooltip-content', '')
    })

    it('should be enabled if query does not contain context filter', () => {
        render(<SearchContextDropdown {...defaultProps} query="test (repo:foo or repo:python)" />)
        expect(screen.getByTestId('dropdown-toggle')).toBeEnabled()
        expect(screen.getByTestId('dropdown-toggle')).toHaveAttribute('data-test-tooltip-content', '')
    })

    it('should be disabled if query contains context filter', () => {
        render(<SearchContextDropdown {...defaultProps} query="test (context:foo or repo:python)" />)
        expect(screen.getByTestId('dropdown-toggle')).toBeDisabled()
        expect(screen.getByTestId('dropdown-toggle')).toHaveAttribute(
            'data-test-tooltip-content',
            'Overridden by query'
        )
    })

    it('should submit search on item click', () => {
        const submitSearch = sinon.spy()

        render(<SearchContextDropdown {...defaultProps} submitSearch={submitSearch} query="test" />)

        userEvent.click(screen.getByTestId('dropdown-toggle'))

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        userEvent.click(screen.getByTestId('search-context-menu-item'))

        sinon.assert.calledOnce(submitSearch)
    })

    it('should close menu when pressing Escape button', () => {
        const closeMenu = sinon.spy()

        render(
            <SearchContextDropdown
                {...defaultProps}
                selectedSearchContextSpec="@username"
                onEscapeMenuClose={closeMenu}
            />
        )

        userEvent.click(screen.getByTestId('dropdown-toggle'))
        userEvent.type(document.body, '{esc}', {
            // We need to skip, otherwise the close event will trigger due to a "click away"
            skipClick: true,
        })

        act(() => {
            // Wait for the next frame
            clock.tick(50)
        })

        sinon.assert.calledOnce(closeMenu)
    })
})
