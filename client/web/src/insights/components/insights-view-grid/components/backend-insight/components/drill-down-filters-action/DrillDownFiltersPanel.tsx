import Popover from '@reach/popover'
import classnames from 'classnames'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import React, { useCallback, useRef, useState } from 'react'
import FocusLock from 'react-focus-lock'

import { flipRightPosition } from '../../../../../context-menu/utils'
import { FormChangeEvent } from '../../../../../form/hooks/useForm'
import { DrillDownFiltersPanel } from '../drill-down-filters-panel/DrillDownFiltersPanel'
import { DrillDownFilters, DrillDownFiltersMode } from '../drill-down-filters-panel/types'

import styles from './DrillDownFiltersPanel.module.scss'
import { useKeyboard } from './hooks/use-keyboard'
import { useOnClickOutside } from './hooks/use-outside-click'

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
    targetRef: React.RefObject<HTMLElement>
    onFilterChange: (filters: DrillDownFilters) => void
}

export const DrillDownFiltersAction: React.FunctionComponent<DrillDownFiltersProps> = props => {
    const { targetRef, filters, onFilterChange } = props

    const [open, setOpen] = useState(false)
    const targetButtonReference = useRef<HTMLButtonElement>(null)
    const popoverReference = useRef<HTMLDivElement>(null)

    const handleTargetClick = (): void => {
        setOpen(open => !open)
    }

    const handleClickOutside = useCallback((event: Event) => {
        if (!targetButtonReference.current) {
            return
        }

        if (targetButtonReference.current.contains(event.target as Node)) {
            return
        }

        setOpen(false)
    }, [])

    const handleEscapePress = useCallback(() => {
        setOpen(false)
    }, [])

    const handleFilterChange = (event: FormChangeEvent<DrillDownFilters>): void => {
        if (event.valid) {
            onFilterChange(event.values)
        }
    }

    // Catch any outside click of popover element
    useOnClickOutside(popoverReference, handleClickOutside)
    // Close popover on escape
    useKeyboard({ detectKeys: ['Escape'] }, handleEscapePress)

    return (
        <>
            <button
                ref={targetButtonReference}
                type="button"
                className={classnames('btn btn-icon btn-secondary rounded-circle p-1', styles.filterButton, {
                    [styles.filterButtonActive]: hasActiveFilters(filters),
                })}
                onClick={handleTargetClick}
            >
                <FilterOutlineIcon size="1rem" />
            </button>

            {open && (
                <Popover
                    ref={popoverReference}
                    targetRef={targetRef}
                    position={flipRightPosition}
                    className={classnames('dropdown-menu', styles.popover)}
                >
                    <FocusLock returnFocus={true}>
                        <DrillDownFiltersPanel
                            initialFiltersValue={filters}
                            className={classnames(styles.filterPanel)}
                            onFiltersChange={handleFilterChange}
                        />
                    </FocusLock>
                </Popover>
            )}
        </>
    )
}
