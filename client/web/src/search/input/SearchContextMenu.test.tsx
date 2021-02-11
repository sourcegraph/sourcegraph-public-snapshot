import { mount } from 'enzyme'
import React from 'react'
import { DropdownItem, DropdownMenu, UncontrolledDropdown } from 'reactstrap'
import sinon from 'sinon'
import { SearchContextMenu, SearchContextMenuProps } from './SearchContextMenu'

describe('SearchContextMenu', () => {
    const defaultProps: SearchContextMenuProps = {
        availableSearchContexts: [
            {
                __typename: 'SearchContext',
                id: '1',
                spec: 'global',
                autoDefined: true,
                description: 'All repositories on Sourcegraph',
            },
            {
                __typename: 'SearchContext',
                id: '2',
                spec: '@username',
                autoDefined: true,
                description: 'Your repositories on Sourcegraph',
            },
            {
                __typename: 'SearchContext',
                id: '3',
                spec: '@username/test-version-1.5',
                autoDefined: true,
                description: 'Only code in version 1.5',
            },
        ],
        defaultSearchContextSpec: 'global',
        selectedSearchContextSpec: 'global',
        setSelectedSearchContextSpec: () => {},
        closeMenu: () => {},
    }

    it('should select item when clicking on it', () => {
        const setSelectedSearchContextSpec = sinon.spy()

        const root = mount(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu {...defaultProps} setSelectedSearchContextSpec={setSelectedSearchContextSpec} />
                </DropdownMenu>
            </UncontrolledDropdown>
        )
        const item = root.find(DropdownItem).at(1)
        item.simulate('click')

        sinon.assert.calledOnce(setSelectedSearchContextSpec)
        sinon.assert.calledWithExactly(setSelectedSearchContextSpec, '@username')
    })

    it('should reset back to default when clicking on Reset button', () => {
        const setSelectedSearchContextSpec = sinon.spy()
        const closeMenu = sinon.spy()

        const root = mount(
            <UncontrolledDropdown>
                <DropdownMenu>
                    <SearchContextMenu
                        {...defaultProps}
                        setSelectedSearchContextSpec={setSelectedSearchContextSpec}
                        selectedSearchContextSpec="@username"
                        closeMenu={closeMenu}
                    />
                </DropdownMenu>
            </UncontrolledDropdown>
        )
        const button = root.find('.search-context-menu__footer-button').at(0)
        button.simulate('click')

        sinon.assert.calledOnce(setSelectedSearchContextSpec)
        sinon.assert.calledWithExactly(setSelectedSearchContextSpec, 'global')

        sinon.assert.calledOnce(closeMenu)
    })
})
