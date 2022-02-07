import classNames from 'classnames'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import React, { DOMAttributes, useRef } from 'react'

import { Button, Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import { SearchBasedBackendFilters } from '../../../../../../core/types/insight/search-insight'
import { SubmissionResult } from '../../../../../form/hooks/useForm'
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
        <Popover isOpen={isOpen} anchor={popoverTargetRef} onOpenChange={event => onVisibilityChange(event.isOpen)}>
            <PopoverTrigger
                as={Button}
                ref={targetButtonReference}
                variant="icon"
                type="button"
                aria-label={isFiltered ? 'Active filters' : 'Filters'}
                className={classNames('btn-icon p-1', styles.filterButton, {
                    [styles.filterButtonWithOpenPanel]: isOpen,
                    [styles.filterButtonActive]: isFiltered,
                })}
            >
                <FilterOutlineIcon className={styles.filterIcon} size="1rem" />
            </PopoverTrigger>

            <PopoverContent
                constrainToScrollParents={false}
                position={Position.rightStart}
                aria-label="Drill-down filters panel"
                onMouseDown={handleMouseDown}
                className={styles.popover}
            >
                <DrillDownFiltersPanel
                    initialFiltersValue={initialFiltersValue}
                    originalFiltersValue={originalFiltersValue}
                    onFiltersChange={onFilterChange}
                    onFilterSave={onFilterSave}
                    onInsightCreate={onInsightCreate}
                />
            </PopoverContent>
        </Popover>
    )
}
