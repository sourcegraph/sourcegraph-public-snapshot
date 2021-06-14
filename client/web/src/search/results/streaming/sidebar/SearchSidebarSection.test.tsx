import { mount } from 'enzyme'
import React from 'react'
import sinon from 'sinon'

import { Filter } from '@sourcegraph/shared/src/search/stream'

import { FilterLink, getDynamicFilterLinks } from './FilterLink'
import { SearchSidebarSection } from './SearchSidebarSection'

describe('SearchSidebarSection', () => {
    const filters: Filter[] = ['typescript', 'JavaScript', 'c++', 'c', 'c#', 'python', 'ruby', 'haskell', 'java'].map(
        lang => ({
            label: `lang:${lang}`,
            value: `lang:${lang}`,
            count: 10,
            limitHit: true,
            kind: 'lang',
        })
    )

    const onFilterChosen = sinon.stub()

    it('should render all items initially', () => {
        const element = mount(
            <SearchSidebarSection header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        const items = element.find(FilterLink)
        expect(items.length).toBe(9)

        const searchbox = element.find('[data-testid="sidebar-section-search-box"]')
        expect(searchbox.length).toBe(1)
    })

    it('should filter items based on search', () => {
        const element = mount(
            <SearchSidebarSection header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        const searchbox = element.find('[data-testid="sidebar-section-search-box"]')
        searchbox.getDOMNode().setAttribute('value', 'Script')
        searchbox.simulate('change', { currentTarget: searchbox })

        const items = element.find(FilterLink)
        expect(items.length).toBe(2)
    })

    it('should clear search when items change', () => {
        const element = mount(
            <SearchSidebarSection header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        let searchbox = element.find('[data-testid="sidebar-section-search-box"]')
        searchbox.getDOMNode().setAttribute('value', 'Script')
        searchbox.simulate('change', { currentTarget: searchbox })

        element.setProps({ children: getDynamicFilterLinks([filters[0], filters[5], filters[3]], onFilterChosen) })
        element.update()

        const items = element.find(FilterLink)
        expect(items.length).toBe(3)

        searchbox = element.find('[data-testid="sidebar-section-search-box"]')
        const searchFilter = searchbox.getDOMNode().getAttribute('value')
        expect(searchFilter).toBe('')
    })

    it('should not show search if only one item in list', () => {
        const element = mount(
            <SearchSidebarSection header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks([filters[2]], onFilterChosen)}
            </SearchSidebarSection>
        )

        const items = element.find(FilterLink)
        expect(items.length).toBe(1)

        const searchbox = element.find('[data-testid="sidebar-section-search-box"]')
        expect(searchbox.length).toBe(0)
    })

    it('should not show search if showSearch is false', () => {
        const element = mount(
            <SearchSidebarSection header="Dynamic filters">
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        const items = element.find(FilterLink)
        expect(items.length).toBe(9)

        const searchbox = element.find('[data-testid="sidebar-section-search-box"]')
        expect(searchbox.length).toBe(0)
    })
})
