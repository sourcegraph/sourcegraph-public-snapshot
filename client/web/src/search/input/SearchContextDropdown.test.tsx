import { mount } from 'enzyme'
import sinon from 'sinon'
import React from 'react'
import * as H from 'history'
import { of } from 'rxjs'
import { SearchPatternType } from '../../graphql-operations'
import { Dropdown, DropdownItem, DropdownToggle } from 'reactstrap'
import { SearchContextDropdown, SearchContextDropdownProps } from './SearchContextDropdown'
import { MockIntersectionObserver } from '../../../../shared/src/util/MockIntersectionObserver'
import { ISearchContext } from '../../../../shared/src/graphql/schema'
import { mockFetchSearchContexts } from '../../searchContexts/testHelpers'

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

    beforeAll(() => {
        window.IntersectionObserver = MockIntersectionObserver
    })

    afterAll(() => {
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
        const item = element.find(DropdownItem).at(0)
        item.simulate('click')

        sinon.assert.calledOnce(submitSearch)
    })

    it('should not submit search if query is empty', () => {
        const submitSearch = sinon.spy()
        const element = mount(<SearchContextDropdown {...defaultProps} submitSearch={submitSearch} query="" />)
        const item = element.find(DropdownItem).at(0)
        item.simulate('click')

        sinon.assert.notCalled(submitSearch)
    })
})
