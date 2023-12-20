import { FC, useMemo } from 'react'

import classNames from 'classnames'

import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter, succeedScan, updateFilter } from '@sourcegraph/shared/src/search/query/transformer'
import type { Filter as ResultFilter } from '@sourcegraph/shared/src/search/stream'
import { Panel } from '@sourcegraph/wildcard'

import {
    FilterTypeList,
    resolveFilterTypeValue,
    toSearchSyntaxTypeFilter,
} from './components/filter-type-list/FilterTypeList'
import { SearchLangFilters } from './components/lang-filter/SearchLangFilters'
import { CodeFilterRecipes, UtilitiesFilterRecipes } from './components/recipes-lists/RecipesLists'
import { useFilterQuery } from './hooks'
import { SearchFilterType, SearchResultFilters } from './types'

import styles from './NewSearchFilters.module.scss'

const TYPES_TO_FILTERS = {
    [SearchFilterType.Code]: [
        SearchResultFilters.ByLanguage,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ByPath,
        SearchResultFilters.Recipes,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Repositories]: [
        SearchResultFilters.ByLanguage,
        SearchResultFilters.ByMetadata,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Paths]: [
        SearchResultFilters.ByLanguage,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Symbols]: [
        SearchResultFilters.BySymbolKind,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ByPath,
    ],
    [SearchFilterType.Commits]: [
        SearchResultFilters.ByAuthor,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ByCommitDate,
        SearchResultFilters.ArchivedAndForked,
    ],
    [SearchFilterType.Diffs]: [
        SearchResultFilters.ByDiffType,
        SearchResultFilters.ByAuthor,
        SearchResultFilters.ByRepository,
        SearchResultFilters.ArchivedAndForked,
    ],
}

interface NewSearchFiltersProps {
    query: string
    filters?: ResultFilter[]
    className?: string
    onQueryChange: (nextQuery: string) => void
}

export const NewSearchFilters: FC<NewSearchFiltersProps> = props => {
    const { query, filters, className, onQueryChange } = props

    const [filterQuery, setFilterQuery] = useFilterQuery()

    const type = useMemo(() => {
        const tokens = scanSearchQuery(query)

        if (tokens.type === 'success') {
            const filters = tokens.term.filter(token => token.type === 'filter') as Filter[]
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
        <Panel
            defaultSize={250}
            minSize={200}
            position="left"
            storageKey="filter-sidebar"
            ariaLabel="Filters sidebar"
            className={classNames(styles.root, className)}
        >
            <aside className={styles.scrollWrapper}>
                <FilterTypeList value={type} onSelect={handleFilterTypeChange} />

                <SearchLangFilters
                    filterType={FilterType.lang}
                    filters={filters}
                    filterQuery={filterQuery}
                    onFilterQueryChange={setFilterQuery}
                />

                <SearchLangFilters
                    filterType={FilterType.repo}
                    filters={filters}
                    filterQuery={filterQuery}
                    onFilterQueryChange={setFilterQuery}
                />

                <SearchLangFilters
                    filterType={FilterType.file}
                    filters={filters}
                    filterQuery={filterQuery}
                    onFilterQueryChange={setFilterQuery}
                />

                <CodeFilterRecipes values={[]} onChange={console.log} />
                <UtilitiesFilterRecipes filters={[]} filtersQuery="" onFilterChange={console.log} />
            </aside>
        </Panel>
    )
}
