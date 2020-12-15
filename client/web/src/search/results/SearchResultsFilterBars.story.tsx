import { storiesOf } from '@storybook/react'
import React from 'react'
import { SearchFilters } from '../../../../shared/src/api/protocol'
import { WebStory } from '../../components/WebStory'
import { QuickLink } from '../../schema/settings.schema'
import { DynamicSearchFilter, SearchResultsFilterBars, SearchResultsFilterBarsProps } from './SearchResultsFilterBars'

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

const { add } = storiesOf('web/search/results/SearchResultsFilterBars', module).addParameters({
    chromatic: { viewports: [993] },
})

add('empty filters', () => <WebStory>{() => <SearchResultsFilterBars {...defaultProps} />}</WebStory>)

add('some generic filters', () => {
    const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }, { value: 'repogroup:my', name: 'My Repos' }]
    return <WebStory>{() => <SearchResultsFilterBars {...defaultProps} genericFilters={genericFilters} />}</WebStory>
})

add('some generic and extension filters', () => {
    const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }, { value: 'repogroup:my', name: 'My Repos' }]
    const extensionFilters: SearchFilters[] = [{ name: 'Extension filter', value: 'repo:test' }]
    return (
        <WebStory>
            {() => (
                <SearchResultsFilterBars
                    {...defaultProps}
                    genericFilters={genericFilters}
                    extensionFilters={extensionFilters}
                />
            )}
        </WebStory>
    )
})

add('some generic, extension, and repo filters', () => {
    const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }, { value: 'repogroup:my', name: 'My Repos' }]
    const extensionFilters: SearchFilters[] = [{ name: 'Extension filter', value: 'repo:test' }]
    const repoFilters: DynamicSearchFilter[] = [{ value: 'archive:yes' }, { value: 'repo:sourcegraph' }]
    return (
        <WebStory>
            {() => (
                <SearchResultsFilterBars
                    {...defaultProps}
                    genericFilters={genericFilters}
                    extensionFilters={extensionFilters}
                    repoFilters={repoFilters}
                />
            )}
        </WebStory>
    )
})

add('some repo filters', () => {
    const repoFilters: DynamicSearchFilter[] = [
        { value: 'archive:yes', count: 5, limitHit: true },
        { value: 'repo:sourcegraph', count: 3, limitHit: false },
    ]
    return <WebStory>{() => <SearchResultsFilterBars {...defaultProps} repoFilters={repoFilters} />}</WebStory>
})

add('result limit hit, render Show More', () => {
    const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }, { value: 'repogroup:my', name: 'My Repos' }]
    const repoFilters: DynamicSearchFilter[] = [
        { value: 'archive:yes', count: 5, limitHit: true },
        { value: 'repo:sourcegraph', count: 3, limitHit: false },
    ]
    return (
        <WebStory>
            {() => (
                <SearchResultsFilterBars
                    {...defaultProps}
                    resultsLimitHit={true}
                    genericFilters={genericFilters}
                    repoFilters={repoFilters}
                />
            )}
        </WebStory>
    )
})

add('filters with quicklinks', () => {
    const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }]
    const repoFilters: DynamicSearchFilter[] = [{ value: 'archive:yes' }]
    const quicklinks: QuickLink[] = [{ name: 'Home', url: '/' }]

    return (
        <WebStory>
            {() => (
                <SearchResultsFilterBars
                    {...defaultProps}
                    genericFilters={genericFilters}
                    repoFilters={repoFilters}
                    quickLinks={quicklinks}
                />
            )}
        </WebStory>
    )
})

add('some filters selected', () => {
    const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }, { value: 'repogroup:my', name: 'My Repos' }]
    const repoFilters: DynamicSearchFilter[] = [
        { value: 'archive:yes', count: 5, limitHit: true },
        { value: 'repo:sourcegraph', count: 3, limitHit: false },
    ]
    return (
        <WebStory>
            {() => (
                <SearchResultsFilterBars
                    {...defaultProps}
                    navbarSearchQuery="repo:sourcegraph lang:c"
                    resultsLimitHit={true}
                    genericFilters={genericFilters}
                    repoFilters={repoFilters}
                />
            )}
        </WebStory>
    )
})

add('search error, display only quicklinks', () => {
    const genericFilters: DynamicSearchFilter[] = [{ value: 'lang:c' }]
    const repoFilters: DynamicSearchFilter[] = [{ value: 'archive:yes' }]
    const quicklinks: QuickLink[] = [{ name: 'Home', url: '/' }]

    return (
        <WebStory>
            {() => (
                <SearchResultsFilterBars
                    {...defaultProps}
                    searchSucceeded={false}
                    genericFilters={genericFilters}
                    repoFilters={repoFilters}
                    quickLinks={quicklinks}
                />
            )}
        </WebStory>
    )
})
