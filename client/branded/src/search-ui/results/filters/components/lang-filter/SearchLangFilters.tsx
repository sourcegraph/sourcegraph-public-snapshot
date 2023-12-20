import { FC, useMemo } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { upperFirst } from 'lodash'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { findFilters } from '@sourcegraph/shared/src/search/query/query'
import { succeedScan } from '@sourcegraph/shared/src/search/query/transformer'
import { Filter } from '@sourcegraph/shared/src/search/stream'
import { Badge, Button, Icon, H4 } from '@sourcegraph/wildcard'

import styles from './SearchLangFilters.module.scss'

interface SearchLangFiltersProps {
    filterType: FilterType
    filters?: Filter[]
    filterQuery: string
    onFilterQueryChange: (nextQuery: string) => void
}

export const SearchLangFilters: FC<SearchLangFiltersProps> = props => {
    const { filters, filterType, filterQuery, onFilterQueryChange } = props

    const selectedLangFilter = useMemo(() => {
        const langFilters = findFilters(succeedScan(filterQuery), filterType)

        return langFilters.length > 0 ? langFilters[0] : null
    }, [filterQuery])

    const isSelected = (filterValue: string): boolean =>
        filterValue === `${selectedLangFilter?.field?.value}:${selectedLangFilter?.value?.value}`

    const handleFilterClick = (langFilter: string, remove?: boolean) => {
        const updatedQuery = remove ? filterQuery.replace(langFilter, '').trim() : `${filterQuery} ${langFilter}`

        onFilterQueryChange(updatedQuery)
    }

    const langFilters = useMemo<Filter[]>(() => {
        if (selectedLangFilter) {
            const selectedLang = filters?.find(filter => isSelected(filter.value))

            return [
                {
                    count: selectedLang?.count ?? 0,
                    label: selectedLang?.label ?? upperFirst(selectedLangFilter?.value?.value),
                    value: `${selectedLangFilter?.field?.value}:${selectedLangFilter?.value?.value}`,
                } as Filter,
            ]
        }

        return filters?.filter(filter => filter.kind === filterType) ?? []
    }, [filters, selectedLangFilter])

    if (langFilters.length === 0) {
        return null
    }

    return (
        <div className={styles.root}>
            <H4 className="mb-0">By {filterType}</H4>
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
                                    <Badge variant="secondary" className="ml-2 mr-1">
                                        {filter.count}
                                    </Badge>
                                )}
                                {isSelectedFilter && (
                                    <Icon svgPath={mdiClose} aria-hidden={true} className="flex-shrink-0" />
                                )}
                            </Button>
                        </li>
                    )
                })}
            </ul>
        </div>
    )
}
