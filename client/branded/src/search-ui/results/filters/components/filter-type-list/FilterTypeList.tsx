import { FC, ReactNode } from 'react'

import { mdiSourceFork, mdiCodeBraces, mdiFileOutline, mdiPlusMinus, mdiFunction, mdiSourceCommit } from '@mdi/js'
import classNames from 'classnames'

import type { Filter } from '@sourcegraph/shared/src/search/stream'
import { Button, Icon } from '@sourcegraph/wildcard'

import { URLQueryFilter } from '../../hooks'
import { DynamicFilterItem } from '../dynamic-filter/SearchDynamicFilter'
import { DynamicFilterBadge } from '../DynamicFilterBadge'

import filterStyles from './../../NewSearchFilters.module.scss'
import styles from './FilterTypeList.module.scss'

interface SearchFilterTypesProps {
    backendFilters: Filter[]
    selectedFilters: URLQueryFilter[]
    onClick: (filter: URLQueryFilter, remove: boolean) => void
}

export const FilterTypeList: FC<SearchFilterTypesProps> = props => {
    const { backendFilters, selectedFilters, setSelectedFilters } = props

    const filters = STATIC_TYPE_FILTERS.map(staticFilter => {
        const backendFilter = backendFilters.find(
            filter => filter.kind === 'type' && filter.label === staticFilter.label
        )
        const selectedFilter = selectedFilters.find(
            filter => filter.kind === 'type' && filter.label === staticFilter.label
        )
        return {
            filter: {
                value: staticFilter.value,
                label: staticFilter.label,
                count: backendFilter?.count ?? 0,
                exhaustive: backendFilter?.exhaustive ?? false,
                kind: staticFilter.kind,
            },
            selected: selectedFilter !== undefined,
        }
    })

    return (
        <ul className={styles.typeList}>
            {filters.map(({ filter, selected }) => (
                <li>
                    <FilterTypeButton filter={filter} selected={selected} onClick={disabled ? undefined : onClick} />
                </li>
            ))}
        </ul>
    )
}

enum SearchTypeLabel {
    Content = 'Content',
    Repositories = 'Repositories',
    Paths = 'Paths',
    Symbols = 'Symbols',
    Commits = 'Commits',
    Diffs = 'Diffs',
}

const STATIC_TYPE_FILTERS: URLQueryFilter[] = [
    { kind: 'type', label: SearchTypeLabel.Content, value: 'type:file' },
    { kind: 'type', label: SearchTypeLabel.Repositories, value: 'type:repo' },
    { kind: 'type', label: SearchTypeLabel.Paths, value: 'type:path' },
    { kind: 'type', label: SearchTypeLabel.Symbols, value: 'type:symbol' },
    { kind: 'type', label: SearchTypeLabel.Commits, value: 'type:commit' },
    { kind: 'type', label: SearchTypeLabel.Diffs, value: 'type:diff' },
]

const FILTER_TYPE_ICONS = {
    [SearchTypeLabel.Content]: mdiCodeBraces,
    [SearchTypeLabel.Repositories]: mdiSourceFork,
    [SearchTypeLabel.Paths]: mdiFileOutline,
    [SearchTypeLabel.Symbols]: mdiFunction,
    [SearchTypeLabel.Commits]: mdiSourceCommit,
    [SearchTypeLabel.Diffs]: mdiPlusMinus,
}

interface FilterTypeButtonProps {
    filter: Filter
    selected: boolean
    onClick?: (filter: URLQueryFilter, remove?: boolean) => void
}

const FilterTypeButton: FC<FilterTypeButtonProps> = props => {
    const { filter, selected, onClick } = props

    return (
        <Button
            variant={selected ? 'primary' : 'secondary'}
            outline={!selected}
            className={classNames(styles.typeListItem, { [styles.typeListItemSelected]: selected })}
            onClick={() => onClick && onClick(filter, selected)}
        >
            <Icon svgPath={FILTER_TYPE_ICONS[filter.label]} aria-hidden={true} />
            <span className={styles.typeListItemText}>{filter.label}</span>
            <DynamicFilterBadge exhaustive={filter.exhaustive} count={filter.count} />
        </Button>
    )
}
