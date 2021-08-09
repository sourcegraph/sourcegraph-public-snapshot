import classnames from 'classnames'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import React, { useRef } from 'react'
import FocusLock from 'react-focus-lock'
import { UncontrolledPopover } from 'reactstrap'

import { FormChangeEvent } from '../../../../../form/hooks/useForm'
import { DrillDownFiltersPanel } from '../drill-down-filters-panel/DrillDownFiltersPanel'
import { DrillDownFilters, DrillDownFiltersMode } from '../drill-down-filters-panel/types'

import styles from './DrillDownFiltersPanel.module.scss'

const hasActiveFilters = (filters: DrillDownFilters): boolean => {
    switch (filters.mode) {
        case DrillDownFiltersMode.Regex:
            return filters.excludeRepoRegex.trim() !== '' || filters.includeRepoRegex.trim() !== ''
        case DrillDownFiltersMode.Repolist:
            // We don't have the repo list mode support yet
            return false
    }
}

interface DrillDownFiltersProps {
    filters: DrillDownFilters
    onFilterChange: (filters: DrillDownFilters) => void
}

export const DrillDownFiltersAction: React.FunctionComponent<DrillDownFiltersProps> = props => {
    const { filters, onFilterChange } = props

    const targetButtonReference = useRef<HTMLButtonElement>(null)

    const handleFilterChange = (event: FormChangeEvent<DrillDownFilters>): void => {
        if (event.valid) {
            onFilterChange(event.values)
        }
    }

    return (
        <>
            <button
                ref={targetButtonReference}
                type="button"
                className={classnames('btn btn-icon btn-secondary rounded-circle p-1', styles.filterButton, {
                    [styles.filterButtonActive]: hasActiveFilters(filters),
                })}
            >
                <FilterOutlineIcon size="1rem" />
            </button>

            <UncontrolledPopover
                placement="right-start"
                target={targetButtonReference}
                trigger="legacy"
                hideArrow={true}
                fade={false}
                popperClassName="border-0"
            >
                <FocusLock returnFocus={true}>
                    <DrillDownFiltersPanel
                        initialFiltersValue={filters}
                        className={classnames(styles.filterPanel)}
                        onFiltersChange={handleFilterChange}
                    />
                </FocusLock>
            </UncontrolledPopover>
        </>
    )
}
