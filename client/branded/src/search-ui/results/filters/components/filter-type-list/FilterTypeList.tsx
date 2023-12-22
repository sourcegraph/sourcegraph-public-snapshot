import { FC } from 'react'

import {
    mdiBook,
    mdiCodeBrackets,
    mdiFileOutline,
    mdiPlusMinus,
    mdiShapeSquareRoundedPlus,
    mdiSourceCommit,
} from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon } from '@sourcegraph/wildcard'

import { SearchFilterType } from '../../types'

import styles from './FilterTypeList.module.scss'

interface SearchFilterTypesProps {
    value: SearchFilterType
    onSelect: (nextTypeValue: SearchFilterType) => void
}

export const FilterTypeList: FC<SearchFilterTypesProps> = props => {
    const { value, onSelect } = props

    return (
        <ul className={styles.typeList}>
            <li>
                <FilterTypeButton
                    type={SearchFilterType.Code}
                    selected={value === SearchFilterType.Code}
                    onClick={onSelect}
                />
            </li>

            <li>
                <FilterTypeButton
                    type={SearchFilterType.Repositories}
                    selected={value === SearchFilterType.Repositories}
                    onClick={onSelect}
                />
            </li>

            <li>
                <FilterTypeButton
                    type={SearchFilterType.Paths}
                    selected={value === SearchFilterType.Paths}
                    onClick={onSelect}
                />
            </li>

            <li>
                <FilterTypeButton
                    type={SearchFilterType.Symbols}
                    selected={value === SearchFilterType.Symbols}
                    onClick={onSelect}
                />
            </li>

            <li>
                <FilterTypeButton
                    type={SearchFilterType.Commits}
                    selected={value === SearchFilterType.Commits}
                    onClick={onSelect}
                />
            </li>

            <li>
                <FilterTypeButton
                    type={SearchFilterType.Diffs}
                    selected={value === SearchFilterType.Diffs}
                    onClick={onSelect}
                />
            </li>
        </ul>
    )
}

const FILTER_TYPE_ICONS = {
    [SearchFilterType.Code]: mdiCodeBrackets,
    [SearchFilterType.Repositories]: mdiBook,
    [SearchFilterType.Paths]: mdiFileOutline,
    [SearchFilterType.Symbols]: mdiShapeSquareRoundedPlus,
    [SearchFilterType.Commits]: mdiSourceCommit,
    [SearchFilterType.Diffs]: mdiPlusMinus,
}

interface FilterTypeButtonProps {
    type: SearchFilterType
    selected: boolean
    onClick: (filterType: SearchFilterType) => void
}

const FilterTypeButton: FC<FilterTypeButtonProps> = props => {
    const { type, selected, onClick } = props

    return (
        <Button
            variant={selected ? 'primary' : 'secondary'}
            outline={!selected}
            className={classNames(styles.typeListItem, { [styles.typeListItemSelected]: selected })}
            onClick={() => onClick(type)}
        >
            <Icon svgPath={FILTER_TYPE_ICONS[type]} aria-hidden={true} />
            {type}
        </Button>
    )
}

export const resolveFilterTypeValue = (value: string | undefined): SearchFilterType => {
    switch (value) {
        case 'repo': {
            return SearchFilterType.Repositories
        }
        case 'path': {
            return SearchFilterType.Paths
        }
        case 'symbol': {
            return SearchFilterType.Symbols
        }
        case 'commit': {
            return SearchFilterType.Commits
        }
        case 'diff': {
            return SearchFilterType.Diffs
        }

        default: {
            return SearchFilterType.Code
        }
    }
}

export const toSearchSyntaxTypeFilter = (value: SearchFilterType): string => {
    switch (value) {
        case SearchFilterType.Repositories: {
            return 'repo'
        }
        case SearchFilterType.Paths: {
            return 'path'
        }
        case SearchFilterType.Symbols: {
            return 'symbol'
        }
        case SearchFilterType.Commits: {
            return 'commit'
        }
        case SearchFilterType.Diffs: {
            return 'diff'
        }

        default: {
            return ''
        }
    }
}
