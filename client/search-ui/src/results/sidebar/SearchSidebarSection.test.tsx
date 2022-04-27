import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { Filter } from '@sourcegraph/shared/src/search/stream'

import { getDynamicFilterLinks } from './FilterLink'
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
        render(
            <SearchSidebarSection sectionId="id" header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        expect(screen.getAllByTestId('filter-link')).toHaveLength(9)

        expect(screen.getByTestId('sidebar-section-search-box')).toBeInTheDocument()
    })

    it('should filter items based on search', () => {
        render(
            <SearchSidebarSection sectionId="id" header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        userEvent.type(screen.getByTestId('sidebar-section-search-box'), 'Script')

        expect(screen.getAllByTestId('filter-link')).toHaveLength(2)
    })

    it('should clear search when items change', () => {
        const { rerender } = render(
            <SearchSidebarSection sectionId="id" header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        userEvent.type(screen.getByTestId('sidebar-section-search-box'), 'Script')

        rerender(
            <SearchSidebarSection sectionId="id" header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks([filters[0], filters[5], filters[3]], onFilterChosen)}
            </SearchSidebarSection>
        )

        expect(screen.getAllByTestId('filter-link')).toHaveLength(3)

        expect(screen.getByTestId('sidebar-section-search-box')).toHaveValue('')
    })

    it('should not show search if only one item in list', () => {
        render(
            <SearchSidebarSection sectionId="id" header="Dynamic filters" showSearch={true}>
                {getDynamicFilterLinks([filters[2]], onFilterChosen)}
            </SearchSidebarSection>
        )

        expect(screen.getByTestId('filter-link')).toBeInTheDocument()

        expect(screen.queryByTestId('sidebar-section-search-box')).not.toBeInTheDocument()
    })

    it('should not show search if showSearch is false', () => {
        render(
            <SearchSidebarSection sectionId="id" header="Dynamic filters">
                {getDynamicFilterLinks(filters, onFilterChosen)}
            </SearchSidebarSection>
        )

        expect(screen.getAllByTestId('filter-link')).toHaveLength(9)

        expect(screen.queryByTestId('sidebar-section-search-box')).not.toBeInTheDocument()
    })
})
