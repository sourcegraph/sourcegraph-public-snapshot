import { FC, ReactNode, useCallback, useMemo, useState } from 'react'

import { mdiClose, mdiSourceRepository } from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { useExperimentalFeatures } from '@sourcegraph/shared/out/src/settings/settings'
import { stringHuman } from '@sourcegraph/shared/src/search/query/printer'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import type { Filter as QueryFilter } from '@sourcegraph/shared/src/search/query/token'
import { omitFilter, succeedScan } from '@sourcegraph/shared/src/search/query/transformer'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { SymbolKind } from '@sourcegraph/shared/src/symbols/SymbolKind'
import { Badge, Button, Icon, H4, Input, LanguageIcon, Code } from '@sourcegraph/wildcard'

import styles from './SearchDynamicFilter.module.scss'

const MAX_FILTERS_NUMBER = 7

interface SearchDynamicFilterProps {
    /**
     * Specifies which type filter we want to render in this particular
     * filter section, it could be lang filter, repo filter, or file filters.
     */
    filterType: string | string[]

    /**
     * Filter query that contains all filter-like query that were applied by users
     * from filters panel UI.
     */
    filterQuery: string

    /**
     * Specifies alternative filter type for the filter, some filters like file
     * have negate-like nature, for example when we want to exclude files, so
     * in order to find these filters in the URL we have to specify -file as alias
     * because in stream API these filters still have file kind.
     */
    filterAlias?: string | string[]

    /**
     * Be default, filters are not exclusive, this means that you can select more
     * than one filter in the section, exclusive:true means that only one selected
     * filter is possible .
     */
    exclusive?: boolean

    /**
     * When filter is selected, it's automatically rendered as a first element
     * in the filters list, staticFilters: true means that we don't change order
     * of the filters when some of them are selected.
     */
    staticFilters?: boolean

    /**
     * List of streamed filters from search stream API
     */
    filters?: Filter[]

    /** Exposes render API to render some custom filter item in the list */
    renderItem?: (filter: Filter) => ReactNode

    /**
     * It's called whenever user changes (pick/reset) any filters in the filter panel.
     * @param nextQuery
     */
    onFilterQueryChange: (nextQuery: string) => void
}

/**
 * Dynamic filter panel section. It renders dynamically generated filters which
 * come from the search stream API.
 */
export const SearchDynamicFilter: FC<SearchDynamicFilterProps> = props => {
    const {
        filters,
        filterType,
        filterAlias,
        filterQuery,
        staticFilters = false,
        exclusive = false,
        renderItem,
        onFilterQueryChange,
    } = props

    const [showAllFilters, setShowAllFilters] = useState(false)
    const [searchTerm, setSearchTerm] = useState<string>('')

    const filterTypes = useMemo(() => toArray(filterType), [filterType])
    const filterAliases = useMemo(() => toArray(filterAlias ?? ''), [filterAlias])

    // Scan the filter query (which comes from URL param) and extract
    // all appearances of a filter type that we're looking for in the
    const filterQueryFilters = useMemo(() => {
        const typedFilters = filterTypes.flatMap(filterType => findFilters(succeedScan(filterQuery), filterType))
        const aliasedFilters = filterAliases.flatMap(filterAlias => findFilters(succeedScan(filterQuery), filterAlias))

        return [...typedFilters, ...aliasedFilters]
    }, [filterQuery, filterAliases, filterTypes])

    // Compares filters stringified value to match selected filters in URL
    const isSelected = useCallback(
        (filterValue: string): boolean => {
            const filteredFilter = filterQueryFilters.find(selectedFilter => isSameFilter(filterValue, selectedFilter))

            return filteredFilter !== undefined
        },
        [filterQueryFilters]
    )

    const mappedFilters = useMemo<Filter[]>(() => {
        // If there are any selected filters in the filters query
        // include these filters in the filters array even if they are not
        // presented in filters from search stream API. If the filter is in both
        // places (URL and steam API) merged them to avoid duplicates in the UI
        if (filterQueryFilters.length > 0 && !staticFilters) {
            const mappedSelectedFilters = filterQueryFilters.map(selectedFilter => {
                const mappedSelectedFilter = filters?.find(filter => isSameFilter(filter.value, selectedFilter))

                return {
                    count: mappedSelectedFilter?.count ?? 0,
                    label: mappedSelectedFilter?.label ?? upperFirst(selectedFilter?.value?.value),
                    value: stringHuman([selectedFilter]),
                } as Filter
            })

            const otherFilters = filterTypes.flatMap(
                filterType => filters?.filter(filter => filter.kind === filterType && !isSelected(filter.value)) ?? []
            )

            return [...mappedSelectedFilters, ...otherFilters]
        }

        return filterTypes.flatMap(filterType => filters?.filter(filter => filter.kind === filterType) ?? [])
    }, [staticFilters, filterTypes, filters, filterQueryFilters])

    const handleFilterClick = (filter: string, remove?: boolean) => {
        const preparedFilterQuery = exclusive
            ? filterQueryFilters.reduce((storeQuery, filter) => omitFilter(storeQuery, filter), filterQuery)
            : filterQuery

        const updatedQuery = remove
            ? preparedFilterQuery.replace(filter, '').trim()
            : `${preparedFilterQuery} ${filter}`

        onFilterQueryChange(updatedQuery)
    }

    if (mappedFilters.length === 0) {
        return null
    }

    const filteredFilters = mappedFilters.filter(filter => filter.label.includes(searchTerm))
    const filtersToShow = showAllFilters ? filteredFilters : filteredFilters.slice(0, MAX_FILTERS_NUMBER)

    return (
        <div className={styles.root}>
            <H4 className={styles.heading}>By {filterTypes.join(',')}</H4>

            {mappedFilters.length > MAX_FILTERS_NUMBER && (
                <Input
                    variant="small"
                    value={searchTerm}
                    placeholder={`Filter ${filterTypes.join(',')}`}
                    onChange={event => setSearchTerm(event.target.value)}
                />
            )}

            <ul className={styles.list}>
                {filtersToShow.map(filter => {
                    const isSelectedFilter = isSelected(filter.value)

                    return (
                        <li key={filter.value}>
                            <Button
                                variant={isSelectedFilter ? 'primary' : 'secondary'}
                                outline={!isSelectedFilter}
                                className={classNames(styles.item, { [styles.itemSelected]: isSelectedFilter })}
                                onClick={() => handleFilterClick(filter.value, isSelectedFilter)}
                            >
                                <span className={styles.itemText}>
                                    {renderItem ? renderItem(filter) : filter.label}
                                </span>
                                {filter.count !== 0 && (
                                    <Badge variant="secondary" className="ml-2">
                                        {filter.count}
                                    </Badge>
                                )}
                                {isSelectedFilter && (
                                    <Icon svgPath={mdiClose} aria-hidden={true} className="ml-1 flex-shrink-0" />
                                )}
                            </Button>
                        </li>
                    )
                })}
            </ul>
            {filteredFilters.length > MAX_FILTERS_NUMBER && (
                <Button variant="link" size="sm" onClick={() => setShowAllFilters(!showAllFilters)}>
                    {showAllFilters ? `Show less ${filterType} filters` : `Show all ${filterType} filters`}
                </Button>
            )}
        </div>
    )
}

const isSameFilter = (filterValue: string, filter: QueryFilter): boolean => {
    const constructedFilterValue = stringHuman([filter])

    return filterValue === constructedFilterValue
}

function toArray<T>(item: T | T[]): T[] {
    if (Array.isArray(item)) {
        return item
    }

    return [item]
}

export const languageFilter = (filter: Filter) => {
    const languageExtension = filter.value.split(':')[1] ?? ''

    return (
        <>
            <LanguageIcon language={languageExtension} className={styles.icon} />
            {filter.label}
        </>
    )
}

export const repoFilter = (filter: Filter) => {
    return (
        <>
            <Icon svgPath={mdiSourceRepository} className={styles.icon} aria-hidden={true} />
            {filter.label}
        </>
    )
}

export const commitDateFilter = (filter: Filter) => {
    return (
        <span className={styles.commitDate}>
            {filter.label}
            <Code>{filter.value}</Code>
        </span>
    )
}

export const symbolFilter = (filter: Filter) => {
    const symbolKindTags = useExperimentalFeatures(features => features.symbolKindTags)

    const symbolType = useMemo(() => {
        const parts = filter.value.split('.')
        return parts[parts.length - 1]
    }, [filter])

    return (
        <>
            <SymbolKind
                kind={symbolType.toUpperCase() as any}
                className={styles.icon}
                symbolKindTags={symbolKindTags}
            />
            {filter.label}
        </>
    )
}

export const utilityFilter = (filter: Filter) => (filter.count === 0 ? filter.value : filter.label)
