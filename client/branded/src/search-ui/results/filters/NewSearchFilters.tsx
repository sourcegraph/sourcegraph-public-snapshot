import { FC, ReactNode, useMemo } from 'react'

import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery, succeedScan } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter as QueryFilter } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import type { Filter } from '@sourcegraph/shared/src/search/stream'
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
import {
    FilterTypeList,
    resolveFilterTypeValue,
    toSearchSyntaxTypeFilter,
} from './components/filter-type-list/FilterTypeList'
import { FiltersDocFooter } from './components/filters-doc-footer/FiltersDocFooter'
import { ArrowBendIcon } from './components/Icons'
import { mergeQueryAndFilters, URLQueryFilter, useUrlFilters } from './hooks'
import { FiltersType, SEARCH_TYPES_TO_FILTER_TYPES, SearchFilterType } from './types'

import styles from './NewSearchFilters.module.scss'

interface NewSearchFiltersProps {
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
}) => {
    const [selectedFilters, setSelectedFilters, serializeFiltersURL] = useUrlFilters()

    const type = useMemo(() => {
        const tokens = scanSearchQuery(query)

        if (tokens.type === 'success') {
            const filters = tokens.term.filter((token): token is QueryFilter => token.type === 'filter')
            const typeFilters = filters.filter(filter => resolveFilter(filter.field.value)?.type === 'type')

            if (typeFilters.length === 0 || typeFilters.length > 1) {
                return SearchFilterType.Code
            }

            return resolveFilterTypeValue(typeFilters[0].value?.value)
        }

        return SearchFilterType.Code
    }, [query])

    const handleFilterTypeChange = (filterType: SearchFilterType): void => {
        const newQuery = changeSearchFilterType(query, filterType)
        const newSelectedFilters = omitImpossibleFilters(selectedFilters, filterType)

        // Replace: true is needed here to avoid populating history with
        // extra entries with completely internal locations update,
        // Setting filters shouldn't be in the history since onQueryChange
        // changes URL itself.
        onQueryChange(newQuery, serializeFiltersURL(newSelectedFilters))
    }

    const handleApplyButtonFilters = (): void => {
        onQueryChange(mergeQueryAndFilters(query, selectedFilters), serializeFiltersURL([]))
    }

    return (
        <div className={styles.scrollWrapper}>
            <FilterTypeList value={type} onSelect={handleFilterTypeChange} />
            <div className={styles.filters}>
                <SearchDynamicFilter
                    title="By repositories"
                    filterKind={FiltersType.Repository}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={repoFilter}
                    onSelectedFilterChange={setSelectedFilters}
                />

                <SearchDynamicFilter
                    title="By language"
                    filterKind={FiltersType.Language}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={languageFilter}
                    onSelectedFilterChange={setSelectedFilters}
                />

                <SearchDynamicFilter
                    title="By symbol kind"
                    filterKind={FiltersType.SymbolKind}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={symbolFilter}
                    onSelectedFilterChange={setSelectedFilters}
                />

                <SearchDynamicFilter
                    title="By author"
                    filterKind={FiltersType.Author}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={authorFilter}
                    onSelectedFilterChange={setSelectedFilters}
                />

                <SearchDynamicFilter
                    title="By commit date"
                    filterKind={FiltersType.CommitDate}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={commitDateFilter}
                    onSelectedFilterChange={setSelectedFilters}
                />

                <SearchDynamicFilter
                    title="By file"
                    filterKind={FiltersType.File}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    onSelectedFilterChange={setSelectedFilters}
                />

                <SearchDynamicFilter
                    title="Utility"
                    filterKind={FiltersType.Utility}
                    filters={filters}
                    selectedFilters={selectedFilters}
                    renderItem={utilityFilter}
                    onSelectedFilterChange={setSelectedFilters}
                />

                <SyntheticCountFilter query={query} isLimitHit={withCountAllFilter} onQueryChange={onQueryChange} />
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

interface SyntheticCountFilterProps {
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
    const { query, isLimitHit, onQueryChange } = props

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

    const handleCountAllFilter = (countFilters: URLQueryFilter[]): void => {
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
            filterKind={FiltersType.Count as any}
            filters={STATIC_COUNT_FILTER}
            selectedFilters={selectedCountFilter}
            renderItem={commitDateFilter}
            onSelectedFilterChange={handleCountAllFilter}
        />
    )
}

function changeSearchFilterType(query: string, searchType: SearchFilterType): string {
    switch (searchType) {
        case SearchFilterType.Code: {
            const filters = findFilters(succeedScan(query), FilterType.type)

            return filters.reduce((query, filter) => omitFilter(query, filter), query)
        }
        default: {
            const filters = findFilters(succeedScan(query), FilterType.type)
            const newQuery = filters.reduce((query, filter) => omitFilter(query, filter), query)

            return updateFilter(newQuery, FilterType.type, toSearchSyntaxTypeFilter(searchType))
        }
    }
}

function omitImpossibleFilters(filters: URLQueryFilter[], searchType: SearchFilterType): URLQueryFilter[] {
    const searchTypePossibleFilters = SEARCH_TYPES_TO_FILTER_TYPES[searchType]

    return filters.filter(filter => searchTypePossibleFilters.includes(filter.kind))
}
