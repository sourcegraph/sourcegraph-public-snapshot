import classNames from 'classnames'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import React, { DOMAttributes, useRef } from 'react'

import { Button } from '@sourcegraph/wildcard'

import { SearchBasedBackendFilters } from '../../../../../../core/types/insight/search-insight'
import { flipRightPosition } from '../../../../../context-menu/utils'
import { SubmissionResult } from '../../../../../form/hooks/useForm'
import { Popover } from '../../../../../popover/Popover'
import { hasActiveFilters } from '../drill-down-filters-panel/components/drill-down-filters-form/DrillDownFiltersForm'
import { DrillDownInsightCreationFormValues } from '../drill-down-filters-panel/components/drill-down-insight-creation-form/DrillDownInsightCreationForm'
import { DrillDownFiltersPanel } from '../drill-down-filters-panel/DrillDownFiltersPanel'

import styles from './DrillDownFiltersPanel.module.scss'

interface DrillDownFiltersProps {
    isOpen: boolean
    initialFiltersValue: SearchBasedBackendFilters
    originalFiltersValue: SearchBasedBackendFilters
    popoverTargetRef: React.RefObject<HTMLElement>
    onFilterChange: (filters: SearchBasedBackendFilters) => void
    onFilterSave: (filters: SearchBasedBackendFilters) => void
    onInsightCreate: (values: DrillDownInsightCreationFormValues) => SubmissionResult
    onVisibilityChange: (open: boolean) => void
}

// To prevent grid layout position change animation. Attempts to drag
// the filter panel should not trigger react-grid-layout events.
const handleMouseDown: DOMAttributes<HTMLElement>['onMouseDown'] = event => event.stopPropagation()

export const DrillDownFiltersAction: React.FunctionComponent<DrillDownFiltersProps> = props => {
    const {
        isOpen,
        popoverTargetRef,
        initialFiltersValue,
        originalFiltersValue,
        onVisibilityChange,
        onFilterChange,
        onFilterSave,
        onInsightCreate,
    } = props

    const targetButtonReference = useRef<HTMLButtonElement>(null)
    const isFiltered = hasActiveFilters(initialFiltersValue)

    return (
        <>
            <Button
                ref={targetButtonReference}
                className={classNames('btn-icon p-1', styles.filterButton, {
                    [styles.filterButtonWithOpenPanel]: isOpen,
                    [styles.filterButtonActive]: isFiltered,
                })}
                aria-label={isFiltered ? 'Active filters' : 'Filters'}
                onMouseDown={handleMouseDown}
            >
                <FilterOutlineIcon className={styles.filterIcon} size="1rem" />
            </Button>

            <Popover
                isOpen={isOpen}
                target={targetButtonReference}
                positionTarget={popoverTargetRef}
                position={flipRightPosition}
                aria-label="Drill-down filters panel"
                onVisibilityChange={onVisibilityChange}
                onMouseDown={handleMouseDown}
            >
                <DrillDownFiltersPanel
                    initialFiltersValue={initialFiltersValue}
                    originalFiltersValue={originalFiltersValue}
                    onFiltersChange={onFilterChange}
                    onFilterSave={onFilterSave}
                    onInsightCreate={onInsightCreate}
                />
            </Popover>
        </>
    )
}
