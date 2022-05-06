import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { act } from 'react-dom/test-utils'
// eslint-disable-next-line no-restricted-imports
import { DropdownMenu, UncontrolledDropdown } from 'reactstrap'
import { Observable, of, throwError } from 'rxjs'
import sinon from 'sinon'

import { ListSearchContextsResult, SearchContextFields } from '@sourcegraph/search'
import { ISearchContext } from '@sourcegraph/shared/src/schema'
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
    ] as ISearchContext[])

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
    ].filter(context => !query || context.spec.toLowerCase().includes(query.toLowerCase())) as SearchContextFields[]
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
        closeMenu: () => {},
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

        render(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} selectSearchContextSpec={selectSearchContextSpec} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        const item = screen.getAllByTestId('search-context-menu-item')[1]
        userEvent.click(item)

        sinon.assert.calledOnce(selectSearchContextSpec)
        sinon.assert.calledWithExactly(selectSearchContextSpec, '@username')
    })

    it('should close menu when pressing Escape button', () => {
        const selectSearchContextSpec = sinon.spy()
        const closeMenu = sinon.spy()

        render(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu
                        {...defaultProps}
                        selectSearchContextSpec={selectSearchContextSpec}
                        selectedSearchContextSpec="@username"
                        closeMenu={closeMenu}
                    />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        const button = screen.getAllByTestId('search-context-menu-header-input')[0]
        userEvent.type(button, '{esc}')
        sinon.assert.calledOnce(closeMenu)
    })

    it('should filter list by spec when searching', () => {
        render(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        const searchInput = screen.getByTestId('search-context-menu-header-input')
        // Search by spec
        userEvent.type(searchInput, 'ser')
        act(() => {
            // Wait for debounce
            clock.tick(500)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        expect(items.length).toBe(2)
        expect(items[0]).toHaveTextContent('@username Your repositories on Sourcegraph')
        expect(items[1]).toHaveTextContent('@username/test-version-1.5 Only code in version 1.5')

        expect(items).toMatchSnapshot()
    })

    it('should show message if search does not find anything', () => {
        render(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

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
        render(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

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
        render(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} fetchSearchContexts={errorFetchSearchContexts} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        expect(items[items.length - 1]).toHaveTextContent('Error occured while loading search contexts')
    })

    it('should default to empty array if fetching auto-defined contexts fails', () => {
        const errorFetchAutoDefinedSearchContexts: () => Observable<ISearchContext[]> = () =>
            throwError(new Error('unknown error'))

        render(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu
                        {...defaultProps}
                        fetchAutoDefinedSearchContexts={errorFetchAutoDefinedSearchContexts}
                    />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })

        const items = screen.getAllByTestId('search-context-menu-item')
        // With no auto-defined contexts, the first context should be a user-defined context
        expect(items[0]).toHaveTextContent('@username/test-version-1.5 Only code in version 1.5')
    })
})
