import React, { useState } from 'react'

import { InsightFilters } from '../../../../../../core'
import { FormChangeEvent, SubmissionResult } from '../../../../../form/hooks/useForm'

import {
    DrillDownFiltersForm,
    DrillDownFiltersFormValues,
} from './components/drill-down-filters-form/DrillDownFiltersForm'
import {
    DrillDownInsightCreationForm,
    DrillDownInsightCreationFormValues,
} from './components/drill-down-insight-creation-form/DrillDownInsightCreationForm'

import styles from './DrillDownFiltersPanel.module.scss'

enum DrillDownFiltersStep {
    Filters = 'filters',
    ViewCreation = 'view-creation',
}

export interface DrillDownFiltersPanelProps {
    initialFiltersValue: InsightFilters
    originalFiltersValue: InsightFilters
    onFiltersChange: (filters: InsightFilters) => void
    onFilterSave: (filters: InsightFilters) => SubmissionResult
    onInsightCreate: (values: DrillDownInsightCreationFormValues) => SubmissionResult
}

export const DrillDownFiltersPanel: React.FunctionComponent<DrillDownFiltersPanelProps> = props => {
    const { initialFiltersValue, originalFiltersValue, onFiltersChange, onFilterSave, onInsightCreate } = props

    const handleFilterChange = (event: FormChangeEvent<DrillDownFiltersFormValues>): void => {
        if (event.valid) {
            onFiltersChange(event.values)
        }
    }

    // By default always render filters mode
    const [step, setStep] = useState(DrillDownFiltersStep.Filters)

    if (step === DrillDownFiltersStep.Filters) {
        return (
            <DrillDownFiltersForm
                className={styles.filtersForm}
                initialFiltersValue={initialFiltersValue}
                originalFiltersValue={originalFiltersValue}
                onFiltersChange={handleFilterChange}
                onFilterSave={onFilterSave}
                onCreateInsightRequest={() => setStep(DrillDownFiltersStep.ViewCreation)}
            />
        )
    }

    return (
        <DrillDownInsightCreationForm
            className={styles.filtersViewCreation}
            onCreateInsight={onInsightCreate}
            onCancel={() => setStep(DrillDownFiltersStep.Filters)}
        />
    )
}
