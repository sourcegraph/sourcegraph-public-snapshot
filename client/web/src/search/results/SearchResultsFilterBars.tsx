import React from 'react'
import { SearchFilters } from '../../../../shared/src/api/protocol'
import { QuickLink } from '../../schema/settings.schema'
import { FilterChip } from '../FilterChip'
import { QuickLinks } from '../QuickLinks'

export interface SearchFilterData {
    name?: string

    value: string

    count?: number
    limitHit?: boolean
}

export const SearchResultsFilterBars: React.FunctionComponent<{
    navbarSearchQuery: string
    resultsFound: boolean
    resultsLimitHit: boolean
    filters: SearchFilterData[]
    extensionFilters: SearchFilters[] | undefined
    repoFilters: SearchFilterData[] | undefined
    quickLinks?: QuickLink[] | undefined
    onFilterClick: (value: string) => void
    onShowMoreResultsClick: (value: string) => void
    calculateShowMoreResultsCount: () => number
}> = ({
    navbarSearchQuery,
    resultsFound,
    resultsLimitHit,
    filters,
    extensionFilters,
    repoFilters,
    quickLinks,
    onFilterClick,
    onShowMoreResultsClick,
    calculateShowMoreResultsCount,
}) => (
    <div className="search-results-filter-bars">
        {((resultsFound && filters.length > 0) || extensionFilters) && (
            <div className="search-results-filter-bars__row" data-testid="filters-bar">
                Filters:
                <div className="search-results-filter-bars__filters">
                    {extensionFilters
                        ?.filter(filter => filter.value !== '')
                        .map(filter => (
                            <FilterChip
                                query={navbarSearchQuery}
                                onFilterChosen={onFilterClick}
                                key={filter.name + filter.value}
                                value={filter.value}
                                name={filter.name}
                            />
                        ))}
                    {filters
                        .filter(filter => filter.value !== '')
                        .map(filter => (
                            <FilterChip
                                query={navbarSearchQuery}
                                onFilterChosen={onFilterClick}
                                key={String(filter.name) + filter.value}
                                value={filter.value}
                                name={filter.name}
                            />
                        ))}
                </div>
            </div>
        )}
        {resultsFound && repoFilters && repoFilters.length > 0 && (
            <div className="search-results-filter-bars__row" data-testid="repo-filters-bar">
                Repositories:
                <div className="search-results-filter-bars__filters">
                    {repoFilters.map(filter => (
                        <FilterChip
                            name={filter.name}
                            query={navbarSearchQuery}
                            onFilterChosen={onFilterClick}
                            key={filter.value}
                            value={filter.value}
                            count={filter.count}
                            limitHit={filter.limitHit}
                        />
                    ))}
                    {resultsLimitHit && !/\brepo:/.test(navbarSearchQuery) && (
                        <FilterChip
                            name="Show more"
                            query={navbarSearchQuery}
                            onFilterChosen={onShowMoreResultsClick}
                            key={`count:${calculateShowMoreResultsCount()}`}
                            value={`count:${calculateShowMoreResultsCount()}`}
                            showMore={true}
                        />
                    )}
                </div>
            </div>
        )}
        <QuickLinks
            quickLinks={quickLinks}
            className="search-results-filter-bars__row search-results-filter-bars__quicklinks"
        />
    </div>
)
