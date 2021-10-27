import { mount } from 'enzyme'
import React, { ChangeEvent } from 'react'
import { act } from 'react-dom/test-utils'
import { DropdownItem, DropdownMenu, UncontrolledDropdown } from 'reactstrap'
import { Observable, of, throwError } from 'rxjs'
import sinon from 'sinon'

import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'
import { mockGetUserSearchContextNamespaces } from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/util/MockIntersectionObserver'

import { ListSearchContextsResult, SearchContextFields } from '../../graphql-operations'

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

        const root = mount(
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
        root.update()

        const item = root.find(DropdownItem).at(1)
        item.simulate('click')

        sinon.assert.calledOnce(selectSearchContextSpec)
        sinon.assert.calledWithExactly(selectSearchContextSpec, '@username')
    })

    it('should close menu when pressing Escape button', () => {
        const selectSearchContextSpec = sinon.spy()
        const closeMenu = sinon.spy()

        const root = mount(
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

        const button = root.find('[data-testid="search-context-menu-header-input"]').at(0)
        button.simulate('keydown', { key: 'Escape' })
        sinon.assert.calledOnce(closeMenu)
    })

    it('should filter list by spec when searching', () => {
        const root = mount(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        const searchInput = root.find('input')
        act(() => {
            // Search by spec
            searchInput.invoke('onInput')?.({
                currentTarget: { value: 'ser' },
            } as ChangeEvent<HTMLInputElement>)
            // Wait for debounce
            clock.tick(500)
        })

        root.update()

        const items = root.find(DropdownItem)
        expect(items.length).toBe(2)
        expect(items.at(0).text()).toBe('@username Your repositories on Sourcegraph')
        expect(items.at(1).text()).toBe('@username/test-version-1.5 Only code in version 1.5')

        expect(items).toMatchSnapshot()
    })

    it('should show message if search does not find anything', () => {
        const root = mount(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        const searchInput = root.find('input')

        act(() => {
            // Search by spec
            searchInput.invoke('onInput')?.({
                currentTarget: { value: 'nothing' },
            } as ChangeEvent<HTMLInputElement>)

            // Wait for debounce
            clock.tick(500)
        })

        root.update()

        const items = root.find(DropdownItem)
        expect(items.length).toBe(1)
        expect(items.at(0).text()).toBe('No contexts found')
    })

    it('should not search by description', () => {
        const root = mount(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )

        const searchInput = root.find('input')

        act(() => {
            searchInput.invoke('onInput')?.({
                currentTarget: { value: 'version 1.5' },
            } as ChangeEvent<HTMLInputElement>)
            // Wait for debounce
            clock.tick(500)
        })

        root.update()

        const items = root.find(DropdownItem)
        expect(items.length).toBe(1)
        expect(items.at(0).text()).toBe('No contexts found')
    })

    it('should show error on failed next page load', () => {
        const errorFetchSearchContexts = () => {
            throw new Error('unknown error')
        }
        const root = mount(
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
        root.update()

        const items = root.find(DropdownItem)
        expect(items.at(items.length - 1).text()).toBe('Error occured while loading search contexts')
    })

    it('should default to empty array if fetching auto-defined contexts fails', () => {
        const errorFetchAutoDefinedSearchContexts: () => Observable<ISearchContext[]> = () =>
            throwError(new Error('unknown error'))

        const root = mount(
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
        root.update()

        const items = root.find(DropdownItem)
        // With no auto-defined contexts, the first context should be a user-defined context
        expect(items.at(0).text()).toBe('@username/test-version-1.5 Only code in version 1.5')
    })
})
