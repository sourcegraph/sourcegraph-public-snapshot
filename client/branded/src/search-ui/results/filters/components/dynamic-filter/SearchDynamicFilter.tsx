import { FC, useCallback, useMemo } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { stringHuman } from '@sourcegraph/shared/out/src/search/query/printer'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { succeedScan } from '@sourcegraph/shared/src/search/query/transformer'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { Badge, Button, Icon, H4 } from '@sourcegraph/wildcard'

import styles from './SearchDynamicFilter.module.scss'

interface SearchDynamicFilterProps {
    /**
     * Specifies which type filter we want to render in this particular
     * filter section, it could be lang filter, repo filter, or file filters.
     */
    filterType: FilterType

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
    filterAlias?: string

    /**
     * List of streamed filters from search stream API
     */
    filters?: Filter[]

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
    const { filters, filterType, filterAlias, filterQuery, onFilterQueryChange } = props

    // Scan the filter query (which comes from URL param) and extract
    // all appearances of a filter type that we're looking for in the
    const filterQueryFilters = useMemo(() => {
        const typedFilters = findFilters(succeedScan(filterQuery), filterType)
        const aliasedFilters = filterAlias ? findFilters(succeedScan(filterQuery), filterAlias) : []

        return [...typedFilters, ...aliasedFilters]
    }, [filterQuery, filterAlias, filterType])

    // Compares filters stringified value to match selected filters in URL
    const isSelected = useCallback(
        (filterValue: string): boolean => {
            const filteredFilter = filterQueryFilters.find(selectedFilter => {
                const constructedFilterValue = stringHuman([selectedFilter])

                return filterValue === constructedFilterValue
            })

            return filteredFilter !== undefined
        },
        [filterQueryFilters]
    )

    const mappedFilters = useMemo<Filter[]>(() => {
        // If there are any selected filters in the filters query
        // include these filters in the filters array even if they are not
        // presented in filters from search stream API. If the filter is in both
        // places (URL and steam API) merged them to avoid duplicates in the UI
        if (filterQueryFilters.length > 0) {
            const mappedSelectedFilters = filterQueryFilters.map(selectedFilter => {
                const mappedSelectedFilter = filters?.find(filter => isSelected(filter.value))

                return {
                    count: mappedSelectedFilter?.count ?? 0,
                    label: mappedSelectedFilter?.label ?? upperFirst(selectedFilter?.value?.value),
                    value: stringHuman([selectedFilter]),
                } as Filter
            })

            const otherFilters =
                filters?.filter(filter => filter.kind === filterType && !isSelected(filter.value)) ?? []

            return [...mappedSelectedFilters, ...otherFilters]
        }

        return filters?.filter(filter => filter.kind === filterType) ?? []
    }, [filters, filterQueryFilters])

    const handleFilterClick = (filter: string, remove?: boolean) => {
        const updatedQuery = remove ? filterQuery.replace(filter, '').trim() : `${filterQuery} ${filter}`

        onFilterQueryChange(updatedQuery)
    }

    if (mappedFilters.length === 0) {
        return null
    }

    return (
        <div className={styles.root}>
            <H4 className={styles.heading}>By {filterType}</H4>
            <ul className={styles.list}>
                {mappedFilters.map(filter => {
                    const isSelectedFilter = isSelected(filter.value)

                    return (
                        <li key={filter.value}>
                            <Button
                                variant={isSelectedFilter ? 'primary' : 'secondary'}
                                outline={!isSelectedFilter}
                                className={classNames(styles.item, { [styles.itemSelected]: isSelectedFilter })}
                                onClick={() => handleFilterClick(filter.value, isSelectedFilter)}
                            >
                                <span className={styles.itemText}>{filter.label}</span>
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
        </div>
    )
}
