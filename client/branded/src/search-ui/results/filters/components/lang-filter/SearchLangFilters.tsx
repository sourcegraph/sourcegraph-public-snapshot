import { FC, useMemo } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { stringHuman } from '@sourcegraph/shared/out/src/search/query/printer'
import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { succeedScan } from '@sourcegraph/shared/src/search/query/transformer'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { Badge, Button, Icon, H4 } from '@sourcegraph/wildcard'

import styles from './SearchLangFilters.module.scss'

interface SearchLangFiltersProps {
    filterType: FilterType
    filterQuery: string
    filterAlias?: string
    filters?: Filter[]
    onFilterQueryChange: (nextQuery: string) => void
}

export const SearchLangFilters: FC<SearchLangFiltersProps> = props => {
    const { filters, filterType, filterAlias, filterQuery, onFilterQueryChange } = props

    const filterQueryFilters = useMemo(() => {
        const typedFilters = findFilters(succeedScan(filterQuery), filterType)
        const aliasedFilters = filterAlias ? findFilters(succeedScan(filterQuery), filterAlias) : []

        return [...typedFilters, ...aliasedFilters]
    }, [filterQuery, filterAlias, filterType])

    const isSelected = (filterValue: string): boolean => {
        const filteredFilter = filterQueryFilters.find(selectedFilter => {
            const constructedFilterValue = stringHuman([selectedFilter])

            return filterValue === constructedFilterValue
        })

        return filteredFilter !== undefined
    }

    const langFilters = useMemo<Filter[]>(() => {
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

    const handleFilterClick = (langFilter: string, remove?: boolean) => {
        const updatedQuery = remove ? filterQuery.replace(langFilter, '').trim() : `${filterQuery} ${langFilter}`

        onFilterQueryChange(updatedQuery)
    }

    if (langFilters.length === 0) {
        return null
    }

    return (
        <div className={styles.root}>
            <H4 className="mb-0 ml-1">By {filterType}</H4>
            <ul className={styles.rootList}>
                {langFilters.map(filter => {
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
