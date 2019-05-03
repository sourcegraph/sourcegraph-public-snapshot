import React from 'react'
import { SearchFilters } from '../../../../shared/src/api/protocol'
import * as GQL from '../../../../shared/src/graphql/schema'
import { FilterChip } from '../FilterChip'
import { isSearchResults } from '../helpers'
import { SearchScopeWithOptionalName } from './SearchResults'

export const SearchResultsFilterBars: React.FunctionComponent<{
    navbarSearchQuery: string
    resultsOrError?: GQL.ISearchResults
    filters: SearchScopeWithOptionalName[]
    extensionFilters: SearchFilters[] | undefined
    onFilterClick: (value: string) => void
    onShowMoreResultsClick: (value: string) => void
    calculateShowMoreResultsCount: () => number
}> = ({
    navbarSearchQuery,
    resultsOrError,
    filters,
    extensionFilters,
    onFilterClick,
    onShowMoreResultsClick,
    calculateShowMoreResultsCount,
}) => (
    <>
        {((isSearchResults(resultsOrError) && filters.length > 0) || extensionFilters) && (
            <div className="search-results__filters-bar" data-testid="filters-bar">
                Filters:
                <div className="search-results__filters">
                    {extensionFilters &&
                        extensionFilters
                            .filter(filter => filter.value !== '')
                            .map((filter, i) => (
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
                        .map((filter, i) => (
                            <FilterChip
                                query={navbarSearchQuery}
                                onFilterChosen={onFilterClick}
                                key={filter.name + filter.value}
                                value={filter.value}
                                name={filter.name}
                            />
                        ))}
                </div>
            </div>
        )}
        {isSearchResults(resultsOrError) &&
            resultsOrError.dynamicFilters.filter(filter => filter.kind === 'repo').length > 0 && (
                <div className="search-results__filters-bar" data-testid="repo-filters-bar">
                    Repositories:
                    <div className="search-results__filters">
                        {resultsOrError.dynamicFilters
                            .filter(filter => filter.kind === 'repo' && filter.value !== '')
                            .map((filter, i) => (
                                <FilterChip
                                    name={filter.label}
                                    query={navbarSearchQuery}
                                    onFilterChosen={onFilterClick}
                                    key={filter.value}
                                    value={filter.value}
                                    count={filter.count}
                                    limitHit={filter.limitHit}
                                />
                            ))}
                        {resultsOrError.limitHit && !/\brepo:/.test(navbarSearchQuery) && (
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
    </>
)
