import Popover from '@reach/popover'
import classnames from 'classnames'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import React, { useCallback, useRef } from 'react'
import FocusLock from 'react-focus-lock'

import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { SearchBasedBackendFilters } from '../../../../../../core/types/insight/search-insight'
import { flipRightPosition } from '../../../../../context-menu/utils'
import { SubmissionResult } from '../../../../../form/hooks/useForm'
import { hasActiveFilters } from '../drill-down-filters-panel/components/drill-down-filters-form/DrillDownFiltersForm'
import { DrillDownInsightCreationFormValues } from '../drill-down-filters-panel/components/drill-down-insight-creation-form/DrillDownInsightCreationForm'
import { DrillDownFiltersPanel } from '../drill-down-filters-panel/DrillDownFiltersPanel'

import styles from './DrillDownFiltersPanel.module.scss'
import { useKeyboard } from './hooks/use-keyboard'
import { useOnClickOutside } from './hooks/use-outside-click'

interface DrillDownFiltersProps {
    isOpen: boolean
    settings: Settings
    initialFiltersValue: SearchBasedBackendFilters
    originalFiltersValue: SearchBasedBackendFilters
    popoverTargetRef: React.RefObject<HTMLElement>
    onFilterChange: (filters: SearchBasedBackendFilters) => void
    onFilterSave: (filters: SearchBasedBackendFilters) => void
    onInsightCreate: (values: DrillDownInsightCreationFormValues) => SubmissionResult
    onVisibilityChange: (open: boolean) => void
}

export const DrillDownFiltersAction: React.FunctionComponent<DrillDownFiltersProps> = props => {
    const {
        isOpen,
        settings,
        popoverTargetRef,
        initialFiltersValue,
        originalFiltersValue,
        onVisibilityChange,
        onFilterChange,
        onFilterSave,
        onInsightCreate,
    } = props

    const targetButtonReference = useRef<HTMLButtonElement>(null)
    const popoverReference = useRef<HTMLDivElement>(null)

    const handleTargetClick = (): void => {
        onVisibilityChange(!isOpen)
    }

    const handleClickOutside = useCallback(
        (event: Event) => {
            if (!targetButtonReference.current) {
                return
            }

            if (targetButtonReference.current.contains(event.target as Node)) {
                return
            }

            onVisibilityChange(false)
        },
        [onVisibilityChange]
    )

    const handleEscapePress = useCallback(() => {
        onVisibilityChange(false)
    }, [onVisibilityChange])

    // Catch any outside click of popover element
    useOnClickOutside(popoverReference, handleClickOutside)
    // Close popover on escape
    useKeyboard({ detectKeys: ['Escape'] }, handleEscapePress)

    const isFiltered = hasActiveFilters(initialFiltersValue)

    return (
        <>
            <button
                ref={targetButtonReference}
                type="button"
                className={classnames('btn btn-icon p-1', styles.filterButton, {
                    [styles.filterButtonWithOpenPanel]: isOpen,
                    [styles.filterButtonActive]: isFiltered,
                })}
                aria-label={isFiltered ? 'Active filters button' : 'Filters button'}
                // To prevent grid layout position change animation. Attempts to drag
                // the filter panel should not trigger react-grid-layout events.
                onMouseDown={event => event.stopPropagation()}
                onClick={handleTargetClick}
            >
                <FilterOutlineIcon className={styles.filterIcon} size="1rem" />
            </button>

            {isOpen && (
                <Popover
                    ref={popoverReference}
                    targetRef={popoverTargetRef}
                    position={flipRightPosition}
                    className={classnames('dropdown-menu', styles.popover)}
                    role="dialog"
                    aria-label="Drill-down filters panel"
                    // To prevent grid layout position change animation. Attempts to drag
                    // the filter panel should not trigger react-grid-layout events.
                    onMouseDown={event => event.stopPropagation()}
                >
                    <FocusLock returnFocus={true}>
                        <DrillDownFiltersPanel
                            settings={settings}
                            initialFiltersValue={initialFiltersValue}
                            originalFiltersValue={originalFiltersValue}
                            onFiltersChange={onFilterChange}
                            onFilterSave={onFilterSave}
                            onInsightCreate={onInsightCreate}
                        />
                    </FocusLock>
                </Popover>
            )}
        </>
    )
}
