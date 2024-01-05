import { FC, useMemo } from 'react'

import { FilterType, NegatedFilters, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery, succeedScan } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter as QueryFilter } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import type { Filter } from '@sourcegraph/shared/src/search/stream'

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
import { useFilterQuery } from './hooks'
import { COMMIT_DATE_FILTERS, SearchFilterType, SYMBOL_KIND_FILTERS } from './types'

import styles from './NewSearchFilters.module.scss'

interface NewSearchFiltersProps {
    query: string
    filters?: Filter[]
    onQueryChange: (nextQuery: string) => void
}

export const NewSearchFilters: FC<NewSearchFiltersProps> = ({ query, filters = [], onQueryChange }) => {
    const [filterQuery, setFilterQuery] = useFilterQuery()

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
        switch (filterType) {
            case SearchFilterType.Code: {
                const filters = findFilters(succeedScan(query), FilterType.type)

                const newQuery = filters.reduce((query, filter) => omitFilter(query, filter), query)
                onQueryChange(newQuery)
                break
            }
            default: {
                const filters = findFilters(succeedScan(query), FilterType.type)
                const newQuery = filters.reduce((query, filter) => omitFilter(query, filter), query)

                onQueryChange(updateFilter(newQuery, FilterType.type, toSearchSyntaxTypeFilter(filterType)))
            }
        }
    }

    return (
        <aside className={styles.scrollWrapper}>
            <FilterTypeList value={type} onSelect={handleFilterTypeChange} />

            {type === SearchFilterType.Symbols && (
                <SearchDynamicFilter
                    title="By symbol kind"
                    filterType={FilterType.select}
                    filters={SYMBOL_KIND_FILTERS}
                    exclusive={true}
                    staticFilters={true}
                    filterQuery={filterQuery}
                    renderItem={symbolFilter}
                    onFilterQueryChange={setFilterQuery}
                />
            )}

            {(type === SearchFilterType.Commits || type === SearchFilterType.Diffs) && (
                <>
                    <SearchDynamicFilter
                        title="By author"
                        filterType={FilterType.author}
                        filters={filters}
                        exclusive={true}
                        filterQuery={filterQuery}
                        renderItem={authorFilter}
                        onFilterQueryChange={setFilterQuery}
                    />

                    <SearchDynamicFilter
                        title="By commit date"
                        filterType={[FilterType.after, FilterType.before]}
                        filters={COMMIT_DATE_FILTERS}
                        exclusive={true}
                        staticFilters={true}
                        filterQuery={filterQuery}
                        renderItem={commitDateFilter}
                        onFilterQueryChange={setFilterQuery}
                    />
                </>
            )}

            <SearchDynamicFilter
                title="By language"
                filterType={FilterType.lang}
                filters={filters}
                filterQuery={filterQuery}
                renderItem={languageFilter}
                onFilterQueryChange={setFilterQuery}
            />

            <SearchDynamicFilter
                title="By repositories"
                filterType={FilterType.repo}
                filters={filters}
                filterQuery={filterQuery}
                renderItem={repoFilter}
                onFilterQueryChange={setFilterQuery}
            />

            <SearchDynamicFilter
                title="By file"
                filterType={FilterType.file}
                filterAlias={NegatedFilters.file}
                filters={filters}
                filterQuery={filterQuery}
                onFilterQueryChange={setFilterQuery}
            />

            <SearchDynamicFilter
                title="Utility"
                filterType="utility"
                filterAlias={[FilterType.archived, FilterType.fork]}
                filters={filters}
                filterQuery={filterQuery}
                renderItem={utilityFilter}
                onFilterQueryChange={setFilterQuery}
            />

            <FiltersDocFooter className={styles.footer} />
        </aside>
    )
}
