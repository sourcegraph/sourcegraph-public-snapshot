import { mount } from 'enzyme'
import React from 'react'
import sinon from 'sinon'
import { SearchFilters } from '../../../../shared/src/api/protocol'
import { FilterChip } from '../FilterChip'
import { DynamicSearchFilter, SearchResultsFilterBars, SearchResultsFilterBarsProps } from './SearchResultsFilterBars'

describe('SearchResultsFilterBars', () => {
    const defaultProps: SearchResultsFilterBarsProps = {
        navbarSearchQuery: 'test',
        searchSucceeded: true,
        resultsLimitHit: false,
        genericFilters: [],
        extensionFilters: [],
        repoFilters: [],
        quickLinks: [],
        onFilterClick: () => {},
        onShowMoreResultsClick: () => {},
        calculateShowMoreResultsCount: () => 0,
    }

    it('should call onFilterClick when filter is clicked on generic filter', () => {
        const onFilterClickSpy = sinon.spy((value: string) => {})
        const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }]
        const element = mount(
            <SearchResultsFilterBars
                {...defaultProps}
                genericFilters={genericFilters}
                onFilterClick={onFilterClickSpy}
            />
        )

        const filterChip = element.find(FilterChip)
        filterChip.simulate('click')

        sinon.assert.calledOnce(onFilterClickSpy)
        sinon.assert.calledWith(onFilterClickSpy, 'lang:c')
    })

    it('should call onFilterClick when filter is clicked on extension filter', () => {
        const onFilterClickSpy = sinon.spy((value: string) => {})
        const extensionFilters: SearchFilters[] = [{ name: 'Extension filter', value: 'repo:test' }]
        const element = mount(
            <SearchResultsFilterBars
                {...defaultProps}
                extensionFilters={extensionFilters}
                onFilterClick={onFilterClickSpy}
            />
        )

        const filterChip = element.find(FilterChip)
        filterChip.simulate('click')

        sinon.assert.calledOnce(onFilterClickSpy)
        sinon.assert.calledWith(onFilterClickSpy, 'repo:test')
    })

    it('should call onFilterClick when filter is clicked on repo filter', () => {
        const onFilterClickSpy = sinon.spy((value: string) => {})
        const repoFilters: DynamicSearchFilter[] = [{ value: 'archive:yes' }]
        const element = mount(
            <SearchResultsFilterBars {...defaultProps} repoFilters={repoFilters} onFilterClick={onFilterClickSpy} />
        )

        const filterChip = element.find(FilterChip)
        filterChip.simulate('click')

        sinon.assert.calledOnce(onFilterClickSpy)
        sinon.assert.calledWith(onFilterClickSpy, 'archive:yes')
    })

    it('should call onShowMoreResultsClick when Show More is clicked', () => {
        const calculateShowMoreResultsCount = () => 5
        const onShowMoreResultsClick = sinon.spy((value: string) => {})

        const repoFilters: DynamicSearchFilter[] = [{ value: 'archive:yes' }]
        const element = mount(
            <SearchResultsFilterBars
                {...defaultProps}
                resultsLimitHit={true}
                repoFilters={repoFilters}
                calculateShowMoreResultsCount={calculateShowMoreResultsCount}
                onShowMoreResultsClick={onShowMoreResultsClick}
            />
        )

        const filterChip = element.find(FilterChip).last()
        filterChip.simulate('click')

        sinon.assert.calledOnce(onShowMoreResultsClick)
        sinon.assert.calledWith(onShowMoreResultsClick, 'count:5')
    })

    it('should escape repo filter values containing spaces', () => {
        const onFilterClickSpy = sinon.spy((value: string) => {})
        const repoFilters: DynamicSearchFilter[] = [{ value: 'repo:foo bar baz' }]
        const element = mount(
            <SearchResultsFilterBars {...defaultProps} repoFilters={repoFilters} onFilterClick={onFilterClickSpy} />
        )

        const filterChip = element.find(FilterChip)
        filterChip.simulate('click')

        sinon.assert.calledOnce(onFilterClickSpy)
        sinon.assert.calledWith(onFilterClickSpy, 'repo:foo\\ bar\\ baz')
    })
})
