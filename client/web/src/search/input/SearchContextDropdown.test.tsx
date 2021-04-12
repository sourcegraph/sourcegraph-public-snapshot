import { mount } from 'enzyme'
import * as H from 'history'
import React from 'react'
import { act } from 'react-dom/test-utils'
import { Dropdown, DropdownItem, DropdownToggle } from 'reactstrap'
import { of } from 'rxjs'
import sinon from 'sinon'

import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/util/MockIntersectionObserver'

import { SearchPatternType } from '../../graphql-operations'
import { mockFetchSearchContexts } from '../../searchContexts/testHelpers'

import { SearchContextDropdown, SearchContextDropdownProps } from './SearchContextDropdown'

const mockFetchAutoDefinedSearchContexts = () =>
    of([
        {
            __typename: 'SearchContext',
            id: '1',
            spec: 'global',
            autoDefined: true,
            description: 'All repositories on Sourcegraph',
            repositories: [],
        },
    ] as ISearchContext[])

describe('SearchContextDropdown', () => {
    const defaultProps: SearchContextDropdownProps = {
        query: '',
        showSearchContextManagement: false,
        fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
        fetchSearchContexts: mockFetchSearchContexts,
        defaultSearchContextSpec: '',
        selectedSearchContextSpec: '',
        setSelectedSearchContextSpec: () => {},
        history: H.createMemoryHistory(),
        caseSensitive: true,
        patternType: SearchPatternType.literal,
        versionContext: undefined,
        submitSearch: () => {},
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
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        const button = element.find(Dropdown)
        expect(button.prop('isOpen')).toBe(false)
    })

    it('should open when toggle event happens', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        let button = element.find(Dropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(Dropdown)
        expect(button.prop('isOpen')).toBe(true)
    })

    it('should close if toggle event happens again', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        let button = element.find(Dropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(Dropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(Dropdown)
        expect(button.prop('isOpen')).toBe(false)
    })

    it('should be enabled if query is empty', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(false)
        expect(dropdown.prop('data-tooltip')).toBe('')
    })

    it('should be enabled if query does not contain context filter', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} query="test (repo:foo or repogroup:python)" />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(false)
        expect(dropdown.prop('data-tooltip')).toBe('')
    })

    it('should be disabled if query contains context filter', () => {
        const element = mount(
            <SearchContextDropdown {...defaultProps} query="test (context:foo or repogroup:python)" />
        )
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(true)
        expect(dropdown.prop('data-tooltip')).toBe('Overridden by query')
    })

    it('should submit search on item click', () => {
        const submitSearch = sinon.spy()
        const element = mount(<SearchContextDropdown {...defaultProps} submitSearch={submitSearch} query="test" />)

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })
        element.update()

        const item = element.find(DropdownItem).at(0)
        item.simulate('click')

        sinon.assert.calledOnce(submitSearch)
    })

    it('should not submit search if submitSearchOnSearchContextChange is false', () => {
        const submitSearch = sinon.spy()
        const element = mount(
            <SearchContextDropdown
                {...defaultProps}
                submitSearch={submitSearch}
                submitSearchOnSearchContextChange={false}
            />
        )

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })
        element.update()

        const item = element.find(DropdownItem).at(0)
        item.simulate('click')

        sinon.assert.notCalled(submitSearch)
    })
})
