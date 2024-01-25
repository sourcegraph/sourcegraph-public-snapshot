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
    onQueryChange: (nextQuery: string, updatedSearchURLQuery?: string) => void
    children?: ReactNode
}

export const NewSearchFilters: FC<NewSearchFiltersProps> = ({ query, filters, onQueryChange, children }) => {
    const [selectedFilters, setSelectedFilters, serilizeFiltersURL] = useUrlFilters()

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
        onQueryChange(newQuery, serilizeFiltersURL(newSelectedFilters))
    }

    const handleApplyButtonFilters = (): void => {
        onQueryChange(mergeQueryAndFilters(query, selectedFilters))
    }

    return (
        <div className={styles.scrollWrapper}>
            <FilterTypeList value={type} onSelect={handleFilterTypeChange} />

            <SearchDynamicFilter
                title="By symbol kind"
                filterKind={FiltersType.SymbolKind}
                filters={filters}
                selectedFilters={selectedFilters}
                renderItem={symbolFilter}
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
                title="By author"
                filterKind={FiltersType.Author}
                filters={filters}
                selectedFilters={selectedFilters}
                renderItem={authorFilter}
                onSelectedFilterChange={setSelectedFilters}
            />

            <SearchDynamicFilter
                title="By repositories"
                filterKind={FiltersType.Repository}
                filters={filters}
                selectedFilters={selectedFilters}
                renderItem={repoFilter}
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

            <div className={styles.footerContent}>
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

                <FiltersDocFooter />
            </div>
        </div>
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
