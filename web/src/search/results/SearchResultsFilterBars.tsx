import React from 'react'
import { SearchFilters } from '../../../../shared/src/api/protocol'
import * as GQL from '../../../../shared/src/graphql/schema'
import { QuickLink } from '../../schema/settings.schema'
import { FilterChip } from '../FilterChip'
import { isScopeSelected, isSearchResults } from '../helpers'
import { QuickLinks } from '../QuickLinks'

export interface SearchScopeWithOptionalName {
    name?: string
    value: string
}

export const SearchResultsFilterBars: React.FunctionComponent<{
    navbarSearchQuery: string
    results?: GQL.ISearchResults
    filters: SearchScopeWithOptionalName[]
    extensionFilters: SearchFilters[] | undefined
    quickLinks?: QuickLink[] | undefined
    onFilterClick: (value: string) => void
    onShowMoreResultsClick: (value: string) => void
    calculateShowMoreResultsCount: () => number
}> = ({
    navbarSearchQuery,
    results,
    filters,
    extensionFilters,
    quickLinks,
    onFilterClick,
    onShowMoreResultsClick,
    calculateShowMoreResultsCount,
}) => (
    <div className="search-results-filter-bars">
        {((isSearchResults(results) && filters.length > 0) || extensionFilters) && (
            <div className="search-results-filter-bars__row" data-testid="filters-bar">
                Filters:
                <div className="search-results-filter-bars__filters">
                    {extensionFilters &&
                        extensionFilters
                            .filter(filter => filter.value !== '')
                            .map((filter, i) => (
                                <FilterChip
                                    isSelected={isScopeSelected(navbarSearchQuery, filter.value)}
                                    onFilterChosen={onFilterClick}
                                    key={filter.name + filter.value}
                                    value={filter.value}
                                    name={filter.name}
                                />
                            ))}
                    {filters
                        .filter(filter => filter.value !== '')
                        .map((filter, i) => (
                            <FilterChip
                                isSelected={isScopeSelected(navbarSearchQuery, filter.value)}
                                onFilterChosen={onFilterClick}
                                key={filter.name + filter.value}
                                value={filter.value}
                                name={filter.name}
                            />
                        ))}
                </div>
            </div>
        )}
        {isSearchResults(results) && results.dynamicFilters.filter(filter => filter.kind === 'repo').length > 0 && (
            <div className="search-results-filter-bars__row" data-testid="repo-filters-bar">
                Repositories:
                <div className="search-results-filter-bars__filters">
                    {results.dynamicFilters
                        .filter(filter => filter.kind === 'repo' && filter.value !== '')
                        .map((filter, i) => (
                            <FilterChip
                                name={filter.label}
                                isSelected={isScopeSelected(navbarSearchQuery, filter.value)}
                                onFilterChosen={onFilterClick}
                                key={filter.value}
                                value={filter.value}
                                count={filter.count}
                                limitHit={filter.limitHit}
                            />
                        ))}
                    {results.limitHit && !/\brepo:/.test(navbarSearchQuery) && (
                        <FilterChip
                            name="Show more"
                            isSelected={false}
                            onFilterChosen={onShowMoreResultsClick}
                            key={`count:${calculateShowMoreResultsCount()}`}
                            value={`count:${calculateShowMoreResultsCount()}`}
                            showMore={true}
                        />
                    )}
                </div>
            </div>
        )}
        {quickLinks && (
            <div className="search-results-filter-bars__row" data-testid="quicklinks-bar">
                <div className="search-results-filter-bars__quicklinks">
                    <QuickLinks quickLinks={quickLinks} />
                </div>
            </div>
        )}
    </div>
)
