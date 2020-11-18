import React from 'react'
import { SearchFilters } from '../../../../shared/src/api/protocol'
import { QuickLink } from '../../schema/settings.schema'
import { FilterChip } from '../FilterChip'
import { QuickLinks } from '../QuickLinks'

export interface DynamicSearchFilter {
    name?: string

    value: string

    count?: number
    limitHit?: boolean
}

export interface SearchResultsFilterBarsProps {
    navbarSearchQuery: string
    searchSucceeded: boolean
    resultsLimitHit: boolean
    genericFilters: DynamicSearchFilter[]
    extensionFilters: SearchFilters[] | undefined
    repoFilters?: DynamicSearchFilter[] | undefined
    quickLinks?: QuickLink[] | undefined
    onFilterClick: (value: string) => void
    onShowMoreResultsClick: (value: string) => void
    calculateShowMoreResultsCount: () => number
}

export const SearchResultsFilterBars: React.FunctionComponent<SearchResultsFilterBarsProps> = ({
    navbarSearchQuery,
    searchSucceeded,
    resultsLimitHit,
    genericFilters,
    extensionFilters,
    repoFilters,
    quickLinks,
    onFilterClick,
    onShowMoreResultsClick,
    calculateShowMoreResultsCount,
}) => (
    <div className="search-results-filter-bars">
        {((searchSucceeded && genericFilters.length > 0) || (extensionFilters && extensionFilters.length > 0)) && (
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
                    {genericFilters
                        .filter(filter => filter.value !== '')
                        .map(filter => (
                            <FilterChip
                                query={navbarSearchQuery}
                                onFilterChosen={onFilterClick}
                                key={String(filter.name) + filter.value}
                                value={filter.value}
                                name={filter.name}
                                count={filter.count}
                                limitHit={filter.limitHit}
                            />
                        ))}
                </div>
            </div>
        )}
        {searchSucceeded && repoFilters && repoFilters.length > 0 && (
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
