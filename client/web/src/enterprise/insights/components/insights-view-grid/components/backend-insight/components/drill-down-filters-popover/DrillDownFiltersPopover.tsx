import { type FC, type RefObject, useRef, useState } from 'react'

import { mdiFilterOutline } from '@mdi/js'
import classNames from 'classnames'

import {
    Button,
    Icon,
    Popover,
    PopoverContent,
    PopoverTrigger,
    PopoverTail,
    Position,
    createRectangle,
    type FormChangeEvent,
    type SubmissionResult,
} from '@sourcegraph/wildcard'

import type { InsightFilters } from '../../../../../../core'
import {
    type DrillDownFiltersFormValues,
    DrillDownInsightCreationForm,
    type DrillDownInsightCreationFormValues,
    DrillDownInsightFilters,
    FilterSectionVisualMode,
    hasActiveFilters,
} from '../drill-down-filters-panel'

import styles from './DrillDownFiltersPopover.module.scss'

const POPOVER_TARGET_PADDING = createRectangle(0, 0, 4, 4)
const POPOVER_CONTAINER_PADDING = { top: 58 }

interface DrillDownFiltersPopoverProps {
    isOpen: boolean
    initialFiltersValue: InsightFilters
    originalFiltersValue: InsightFilters
    isNumSamplesFilterAvailable: boolean
    anchor: RefObject<HTMLElement>
    onFilterChange: (filters: InsightFilters) => void
    onFilterSave: (filters: InsightFilters) => void
    onInsightCreate: (values: DrillDownInsightCreationFormValues) => SubmissionResult
    onVisibilityChange: (open: boolean) => void
}

export enum DrillDownFiltersStep {
    Filters = 'filters',
    ViewCreation = 'view-creation',
}

const STEP_STYLES = {
    [DrillDownFiltersStep.Filters]: styles.popoverWithFilters,
    [DrillDownFiltersStep.ViewCreation]: styles.popoverWithViewCreation,
}

export const DrillDownFiltersPopover: FC<DrillDownFiltersPopoverProps> = props => {
    const {
        isOpen,
        anchor,
        initialFiltersValue,
        originalFiltersValue,
        isNumSamplesFilterAvailable,
        onVisibilityChange,
        onFilterChange,
        onFilterSave,
        onInsightCreate,
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
                className={classNames('p-1', styles.filterButton, {
                    [styles.filterButtonWithOpenPanel]: isOpen,
                    [styles.filterButtonActive]: isFiltered,
                })}
            >
                <Icon
                    className={styles.filterIcon}
                    svgPath={mdiFilterOutline}
                    inline={false}
                    aria-hidden={true}
                    height="1rem"
                    width="1rem"
                />
            </PopoverTrigger>

            <PopoverContent
                position={Position.rightStart}
                constrainToScrollParents={true}
                targetPadding={POPOVER_TARGET_PADDING}
                constraintPadding={POPOVER_CONTAINER_PADDING}
                aria-label="Drill-down filters panel"
                className={classNames(styles.popover, STEP_STYLES[step])}
                onKeyDown={event => event.stopPropagation()}
            >
                {step === DrillDownFiltersStep.Filters && (
                    <DrillDownInsightFilters
                        initialValues={initialFiltersValue}
                        originalValues={originalFiltersValue}
                        isNumSamplesFilterAvailable={isNumSamplesFilterAvailable}
                        visualMode={FilterSectionVisualMode.CollapseSections}
                        onFiltersChange={handleFilterChange}
                        onFilterSave={onFilterSave}
                        onCreateInsightRequest={() => setStep(DrillDownFiltersStep.ViewCreation)}
                    />
                )}

                {step === DrillDownFiltersStep.ViewCreation && (
                    <DrillDownInsightCreationForm
                        onCreateInsight={handleCreateInsight}
                        onCancel={() => setStep(DrillDownFiltersStep.Filters)}
                    />
                )}
            </PopoverContent>

            <PopoverTail size="sm" />
        </Popover>
    )
}
