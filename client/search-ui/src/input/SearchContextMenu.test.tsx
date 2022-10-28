import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { act } from 'react-dom/test-utils'
import { Observable, of, throwError } from 'rxjs'
import sinon from 'sinon'

import { ListSearchContextsResult, SearchContextMinimalFields } from '@sourcegraph/search'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/testing/MockIntersectionObserver'
import { mockGetUserSearchContextNamespaces } from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SearchContextMenu, SearchContextMenuProps } from './SearchContextMenu'

const mockFetchAutoDefinedSearchContexts = () =>
    of([
        {
            __typename: 'SearchContext',
            id: '1',
            spec: 'global',
            name: 'global',
            namespace: null,
            autoDefined: true,
            description: 'All repositories on Sourcegraph',
            query: '',
            repositories: [],
            public: true,
            updatedAt: '2021-03-15T19:39:11Z',
            viewerCanManage: false,
        },
        {
            __typename: 'SearchContext',
            id: '2',
            spec: '@username',
            name: 'username',
            namespace: {
                __typename: 'User',
                id: 'u1',
                namespaceName: 'username',
            },
            autoDefined: true,
            description: 'Your repositories on Sourcegraph',
            query: '',
            repositories: [],
            public: true,
            updatedAt: '2021-03-15T19:39:11Z',
            viewerCanManage: false,
        },
    ] as SearchContextMinimalFields[])

const mockFetchSearchContexts = ({ query }: { first: number; query?: string; after?: string }) => {
    const nodes = [
        {
            __typename: 'SearchContext',
            id: '3',
            spec: '@username/test-version-1.5',
            name: 'test-version-1.5',
            namespace: {
                __typename: 'User',
                id: 'u1',
                namespaceName: 'username',
            },
            autoDefined: false,
            public: true,
            description: 'Only code in version 1.5',
            query: '',
            updatedAt: '2021-03-15T19:39:11Z',
            repositories: [],
            viewerCanManage: true,
        },
        {
            __typename: 'SearchContext',
            id: '4',
            spec: '@org/test-version-1.6',
            name: 'test-version-1.6',
            namespace: {
                __typename: 'Org',
                id: 'o1',
                namespaceName: 'org',
            },
            autoDefined: false,
            public: true,
            description: 'Only code in version 1.6',
            query: '',
            updatedAt: '2021-03-15T19:39:11Z',
            repositories: [],
            viewerCanManage: true,
        },
    ].filter(
        context => !query || context.spec.toLowerCase().includes(query.toLowerCase())
    ) as SearchContextMinimalFields[]
    const result: ListSearchContextsResult['searchContexts'] = {
        nodes,
        pageInfo: {
            endCursor: 'foo',
            hasNextPage: false,
        },
        totalCount: nodes.length,
    }
    return of(result)
}

describe('SearchContextMenu', () => {
    const defaultProps: SearchContextMenuProps = {
        authenticatedUser: null,
        showSearchContextManagement: false,
        defaultSearchContextSpec: 'global',
        selectedSearchContextSpec: 'global',
        selectSearchContextSpec: () => {},
        fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts,
        fetchSearchContexts: mockFetchSearchContexts,
        onMenuClose: () => {},
        getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
        searchContextsEnabled: true,
        platformContext: NOOP_PLATFORM_CONTEXT,
        telemetryService: NOOP_TELEMETRY_SERVICE,
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

    it('should select item when clicking on it', () => {
        const selectSearchContextSpec = sinon.spy()

        render(<SearchContextMenu {...defaultProps} selectSearchContextSpec={selectSearchContextSpec} />)

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        const item = screen.getAllByTestId('search-context-menu-item')[1]
        userEvent.click(item)

        sinon.assert.calledOnce(selectSearchContextSpec)
        sinon.assert.calledWithExactly(selectSearchContextSpec, '@username')
    })

    it('should filter list by spec when searching', () => {
        render(<SearchContextMenu {...defaultProps} />)

        const searchInput = screen.getByTestId('search-context-menu-header-input')
        // Search by spec
        userEvent.type(searchInput, 'ser')
        act(() => {
            // Wait for debounce
            clock.tick(500)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        expect(items.length).toBe(2)
        expect(items[0]).toHaveTextContent('@username, Your repositories on Sourcegraph')
        expect(items[1]).toHaveTextContent('@username/test-version-1.5, Only code in version 1.5')

        expect(items).toMatchSnapshot()
    })

    it('should show message if search does not find anything', () => {
        render(<SearchContextMenu {...defaultProps} />)

        const searchInput = screen.getByTestId('search-context-menu-header-input')
        // Search by spec
        userEvent.type(searchInput, 'nothing')
        act(() => {
            // Wait for debounce
            clock.tick(500)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        expect(items[0]).toHaveTextContent('No contexts found')
    })

    it('should not search by description', () => {
        render(<SearchContextMenu {...defaultProps} />)

        const searchInput = screen.getByTestId('search-context-menu-header-input')
        userEvent.type(searchInput, 'version 1.5')
        act(() => {
            // Wait for debounce
            clock.tick(500)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        expect(items[0]).toHaveTextContent('No contexts found')
    })

    it('should show error on failed next page load', () => {
        const errorFetchSearchContexts = () => {
            throw new Error('unknown error')
        }
        render(<SearchContextMenu {...defaultProps} fetchSearchContexts={errorFetchSearchContexts} />)

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        expect(items[items.length - 1]).toHaveTextContent('Error occurred while loading search contexts')
    })

    it('should default to empty array if fetching auto-defined contexts fails', () => {
        const errorFetchAutoDefinedSearchContexts: () => Observable<SearchContextMinimalFields[]> = () =>
            throwError(new Error('unknown error'))

        render(
            <SearchContextMenu {...defaultProps} fetchAutoDefinedSearchContexts={errorFetchAutoDefinedSearchContexts} />
        )

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        // With no auto-defined contexts, the first context should be a user-defined context
        expect(items[0]).toHaveTextContent('@username/test-version-1.5, Only code in version 1.5')
    })
})
