import { FC, ReactNode, useEffect, useCallback, useMemo } from 'react'

import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery, succeedScan } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter as QueryFilter } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter } from '@sourcegraph/shared/src/search/query/transformer'
import type { Filter } from '@sourcegraph/shared/src/search/stream'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import {
    authorFilter,
    commitDateFilter,
    languageFilter,
    repoFilter,
    SearchDynamicFilter,
    symbolFilter,
    utilityFilter,
} from './components/dynamic-filter/SearchDynamicFilter'
import { FilterTypeList } from './components/filter-type-list/FilterTypeList'
import { FiltersDocFooter } from './components/filters-doc-footer/FiltersDocFooter'
import { ArrowBendIcon } from './components/Icons'
import { mergeQueryAndFilters, URLQueryFilter, useUrlFilters } from './hooks'
import { FilterKind, SearchTypeLabel, SEARCH_TYPES_TO_FILTER_TYPES } from './types'

import styles from './NewSearchFilters.module.scss'

interface NewSearchFiltersProps extends TelemetryProps {
    query: string
    filters?: Filter[]
    withCountAllFilter: boolean
    onQueryChange: (nextQuery: string, updatedSearchURLQuery?: string) => void
    children?: ReactNode
}

export const NewSearchFilters: FC<NewSearchFiltersProps> = ({
    query,
    filters,
    withCountAllFilter,
    onQueryChange,
    children,
    telemetryService,
}) => {
    const [selectedFilters, setSelectedFilters, serializeFiltersURL] = useUrlFilters()

    useEffect(() => {
        if (queryHasTypeFilter(query) && selectedFilters.some(filter => filter.kind === 'type')) {
            setSelectedFilters(selectedFilters.filter(filter => filter.kind !== 'type'))
        }
    }, [selectedFilters, query, setSelectedFilters])

    const handleFilterTypeClick = useCallback(
        (filter: URLQueryFilter, remove: boolean): void => {
            telemetryService.log('SearchFiltersTypeClick', { filterType: filter.label }, { filterType: filter.label })
            if (remove) {
                setSelectedFilters(
                    selectedFilters.filter(
                        selectedFilter => selectedFilter.kind !== 'type' || selectedFilter.label !== filter.label
                    )
                )
            } else {
                const relevantFilters = omitImpossibleFilters(selectedFilters, filter.label as SearchTypeLabel)
                setSelectedFilters([
                    ...relevantFilters.filter(relevantFilters => relevantFilters.kind !== 'type'),
                    filter,
                ])
            }
        },
        [selectedFilters, setSelectedFilters, telemetryService]
    )

    const handleFilterChange = useCallback(
        (filterKind: Filter['kind'], filters: URLQueryFilter[]) => {
            setSelectedFilters(filters)
            telemetryService.log('SearchFiltersSelectFilter', { filterKind }, { filterKind })
        },
        [setSelectedFilters, telemetryService]
    )

    const handleApplyButtonFilters = (): void => {
        onQueryChange(mergeQueryAndFilters(query, selectedFilters), serializeFiltersURL([]))
        telemetryService.log('SearchFiltersApplyFiltersClick')
    }

    return (
        <div className={styles.scrollWrapper}>
            <FilterTypeList
                backendFilters={filters ?? []}
                disabled={queryHasTypeFilter(query)}
                selectedFilters={selectedFilters}
                onClick={handleFilterTypeClick}
            />
            <div className={styles.filters}>
                <SearchDynamicFilter
                    title="By repositories"
                    filterKind={FilterKind.Repository}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={repoFilter}
                    onSelectedFilterChange={handleFilterChange}
                />

                <SearchDynamicFilter
                    title="By language"
                    filterKind={FilterKind.Language}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={languageFilter}
                    onSelectedFilterChange={handleFilterChange}
                />

                <SearchDynamicFilter
                    title="By symbol kind"
                    filterKind={FilterKind.SymbolKind}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={symbolFilter}
                    onSelectedFilterChange={handleFilterChange}
                />

                <SearchDynamicFilter
                    title="By author"
                    filterKind={FilterKind.Author}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={authorFilter}
                    onSelectedFilterChange={handleFilterChange}
                />

                <SearchDynamicFilter
                    title="By commit date"
                    filterKind={FilterKind.CommitDate}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={commitDateFilter}
                    onSelectedFilterChange={handleFilterChange}
                />

                <SearchDynamicFilter
                    title="By file"
                    filterKind={FilterKind.File}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    onSelectedFilterChange={handleFilterChange}
                />

                <SearchDynamicFilter
                    title="Utility"
                    filterKind={FilterKind.Utility}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={utilityFilter}
                    onSelectedFilterChange={handleFilterChange}
                />

                <SyntheticCountFilter
                    query={query}
                    isLimitHit={withCountAllFilter}
                    onQueryChange={onQueryChange}
                    telemetryService={telemetryService}
                />
            </div>

            <footer className={styles.actions}>
                {selectedFilters.length > 0 && (
                    <Tooltip
                        placement="right"
                        content="Moves all your applied filters from this panel into the query bar at the top and resets selected options from this panel."
                    >
                        <Button variant="secondary" outline={true} onClick={handleApplyButtonFilters}>
                            Move filters to the query
                            <Icon as={ArrowBendIcon} aria-hidden={true} className={styles.moveIcon} />
                        </Button>
                    </Tooltip>
                )}

                {children}
            </footer>

            <FiltersDocFooter className={styles.footerDoc} />
        </div>
    )
}

function queryHasTypeFilter(query: string): boolean {
    const tokens = scanSearchQuery(query)
    if (tokens.type !== 'success') {
        return false
    }
    const filters = tokens.term.filter((token): token is QueryFilter => token.type === 'filter')
    return filters.some(filter => filter.field.value === 'type')
}

const STATIC_COUNT_FILTER: Filter[] = [
    {
        // Since backend filters don't support count filter
        // this means that Filter type doesn't have this as possible kind value
        // It's okay to cast it since it's handled in one place below in SyntheticCountFilter
        kind: 'count' as any,
        label: 'Show all matches',
        count: 0,
        exhaustive: true,
        value: 'count:all',
    },
]

interface SyntheticCountFilterProps extends TelemetryProps {
    query: string
    isLimitHit: boolean
    onQueryChange: (query: string) => void
}

/**
 * Client-based count filter, usually filters are generated on the backend
 * based on a given query and results, count filter isn't supported by our
 * backend at the moment, so it's synthetically included on the client.
 *
 * It scans original search query and detects count filter if it catches
 * count:all filter. As user enables this filter in the filter panel it
 * changes the original query by adding count:all to the end.
 */
const SyntheticCountFilter: FC<SyntheticCountFilterProps> = props => {
    const { query, isLimitHit, onQueryChange, telemetryService } = props

    const selectedCountFilter = useMemo<Filter[]>(() => {
        const tokens = scanSearchQuery(query)

        if (tokens.type === 'success') {
            const filters = tokens.term.filter((token): token is QueryFilter => token.type === 'filter')
            const countFilters = filters.filter(filter => resolveFilter(filter.field.value)?.type === 'count')

            if (countFilters.length === 0 || countFilters.length > 1) {
                return []
            }

            return STATIC_COUNT_FILTER
        }

        return []
    }, [query])

    const handleCountAllFilter = (filterKind: Filter['kind'], countFilters: URLQueryFilter[]): void => {
        telemetryService.log('SearchFiltersSelectFilter', { filterKind }, { filterKind })

        if (countFilters.length > 0) {
            onQueryChange(`${query} count:all`)
        } else {
            const filters = findFilters(succeedScan(query), FilterType.count)
            const nextQuery = filters.reduce((query, filter) => omitFilter(query, filter), query)

            onQueryChange(nextQuery)
        }
    }

    // Hide count all filter if search is already exhaustive
    if (selectedCountFilter.length === 0 && !isLimitHit) {
        return null
    }

    return (
        <SearchDynamicFilter
            filterKind={FilterKind.Count as any}
            filters={STATIC_COUNT_FILTER}
            selectedFilters={selectedCountFilter}
            renderItem={commitDateFilter}
            onSelectedFilterChange={handleCountAllFilter}
        />
    )
}

function omitImpossibleFilters(filters: URLQueryFilter[], searchType: SearchTypeLabel): URLQueryFilter[] {
    const searchTypePossibleFilters = SEARCH_TYPES_TO_FILTER_TYPES[searchType]
    return filters.filter(filter => searchTypePossibleFilters.includes(filter.kind))
}
