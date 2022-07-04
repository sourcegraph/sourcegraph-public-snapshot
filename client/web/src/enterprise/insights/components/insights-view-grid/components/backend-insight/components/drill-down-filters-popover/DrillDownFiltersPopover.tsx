import React, { DOMAttributes, useRef, useState } from 'react'

import classNames from 'classnames'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'

import { Button, createRectangle, Popover, PopoverContent, PopoverTrigger, Position } from '@sourcegraph/wildcard'

import { SeriesDisplayOptionsInput } from '../../../../../../../../graphql-operations'
import { Insight, InsightFilters } from '../../../../../../core'
import { SeriesDisplayOptions, SeriesDisplayOptionsInputRequired } from '../../../../../../core/types/insight/common'
import { FormChangeEvent, SubmissionResult } from '../../../../../form/hooks/useForm'
import {
    DrillDownInsightCreationForm,
    DrillDownInsightCreationFormValues,
    DrillDownFiltersFormValues,
    DrillDownInsightFilters,
    FilterSectionVisualMode,
    hasActiveFilters,
} from '../drill-down-filters-panel'
import { parseSeriesDisplayOptions } from '../drill-down-filters-panel/drill-down-filters/utils'

import styles from './DrillDownFiltersPopover.module.scss'

const POPOVER_PADDING = createRectangle(0, 0, 5, 5)
interface DrillDownFiltersPopoverProps {
    isOpen: boolean
    initialFiltersValue: InsightFilters
    originalFiltersValue: InsightFilters
    anchor: React.RefObject<HTMLElement>
    insight: Insight
    onFilterChange: (filters: InsightFilters) => void
    onFilterSave: (filters: InsightFilters, displayOptions: SeriesDisplayOptionsInput) => void
    onInsightCreate: (values: DrillDownInsightCreationFormValues) => SubmissionResult
    onVisibilityChange: (open: boolean) => void
    originalSeriesDisplayOptions: SeriesDisplayOptions
    onSeriesDisplayOptionsChange: (options: SeriesDisplayOptionsInputRequired) => void
}

// To prevent grid layout position change animation. Attempts to drag
// the filter panel should not trigger react-grid-layout events.
const handleMouseDown: DOMAttributes<HTMLElement>['onMouseDown'] = event => event.stopPropagation()

export enum DrillDownFiltersStep {
    Filters = 'filters',
    ViewCreation = 'view-creation',
}

const STEP_STYLES = {
    [DrillDownFiltersStep.Filters]: styles.popoverWithFilters,
    [DrillDownFiltersStep.ViewCreation]: styles.popoverWithViewCreation,
}

export const DrillDownFiltersPopover: React.FunctionComponent<
    React.PropsWithChildren<DrillDownFiltersPopoverProps>
> = props => {
    const {
        isOpen,
        anchor,
        initialFiltersValue,
        originalFiltersValue,
        onVisibilityChange,
        onFilterChange,
        onFilterSave,
        onInsightCreate,
        originalSeriesDisplayOptions,
        onSeriesDisplayOptionsChange,
    } = props

    // By default always render filters mode
    const [step, setStep] = useState(DrillDownFiltersStep.Filters)
    const targetButtonReference = useRef<HTMLButtonElement>(null)
    const isFiltered = hasActiveFilters(initialFiltersValue)

    const handleFilterChange = (event: FormChangeEvent<DrillDownFiltersFormValues>): void => {
        if (event.valid) {
            onFilterChange(event.values)
        }
    }

    const handleCreateInsight = (values: DrillDownInsightCreationFormValues): void => {
        setStep(DrillDownFiltersStep.Filters)
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        onInsightCreate(values)
    }

    return (
        <Popover isOpen={isOpen} anchor={anchor} onOpenChange={event => onVisibilityChange(event.isOpen)}>
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
                targetPadding={POPOVER_PADDING}
                constrainToScrollParents={true}
                position={Position.rightStart}
                aria-label="Drill-down filters panel"
                onMouseDown={handleMouseDown}
                className={classNames(styles.popover, STEP_STYLES[step])}
            >
                {step === DrillDownFiltersStep.Filters && (
                    <DrillDownInsightFilters
                        initialValues={initialFiltersValue}
                        originalValues={originalFiltersValue}
                        visualMode={FilterSectionVisualMode.CollapseSections}
                        onFiltersChange={handleFilterChange}
                        onFilterSave={onFilterSave}
                        onCreateInsightRequest={() => setStep(DrillDownFiltersStep.ViewCreation)}
                        originalSeriesDisplayOptions={parseSeriesDisplayOptions(originalSeriesDisplayOptions)}
                        onSeriesDisplayOptionsChange={onSeriesDisplayOptionsChange}
                    />
                )}

                {step === DrillDownFiltersStep.ViewCreation && (
                    <DrillDownInsightCreationForm
                        onCreateInsight={handleCreateInsight}
                        onCancel={() => setStep(DrillDownFiltersStep.Filters)}
                    />
                )}
            </PopoverContent>
        </Popover>
    )
}
