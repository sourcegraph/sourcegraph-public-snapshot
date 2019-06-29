import LinkIcon from 'mdi-react/LinkIcon'
import React from 'react'
import { SearchFilters } from '../../../../shared/src/api/protocol'
import * as GQL from '../../../../shared/src/graphql/schema'
import { QuickLink } from '../../schema/settings.schema'
import { FilterChip } from '../FilterChip'
import { isSearchResults } from '../helpers'

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
        {isSearchResults(results) && results.dynamicFilters.filter(filter => filter.kind === 'repo').length > 0 && (
            <div className="search-results-filter-bars__row" data-testid="repo-filters-bar">
                Repositories:
                <div className="search-results-filter-bars__filters">
                    {results.dynamicFilters
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
                    {results.limitHit && !/\brepo:/.test(navbarSearchQuery) && (
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
        {quickLinks && (
            <div className="search-results-filter-bars__row" data-testid="quicklinks-bar">
                <div className="search-results-filter-bars__filters search-results-filter-bars__filters--no-label">
                    {quickLinks.map((quickLink, i) => (
                        <small className="search-results-filter-bars__filters-quicklink text-nowrap">
                            <a href={quickLink.url} data-tooltip={quickLink.description} key={i}>
                                <LinkIcon className="icon-inline pr-1" />
                                {quickLink.name}
                            </a>
                        </small>
                    ))}
                </div>
            </div>
        )}
    </div>
)
