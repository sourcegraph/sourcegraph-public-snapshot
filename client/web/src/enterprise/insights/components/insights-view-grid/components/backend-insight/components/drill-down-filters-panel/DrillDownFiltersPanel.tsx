import React, { useState } from 'react'

import { SearchBasedBackendFilters } from '../../../../../../core/types/insight/search-insight'
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
    initialFiltersValue: SearchBasedBackendFilters
    originalFiltersValue: SearchBasedBackendFilters
    onFiltersChange: (filters: SearchBasedBackendFilters) => void
    onFilterSave: (filters: SearchBasedBackendFilters) => SubmissionResult
    onInsightCreate: (values: DrillDownInsightCreationFormValues) => SubmissionResult
}

export const DrillDownFiltersPanel: React.FunctionComponent<DrillDownFiltersPanelProps> = props => {
    const {
        initialFiltersValue,
        originalFiltersValue,
        onFiltersChange,
        onFilterSave,
        onInsightCreate,
    } = props

    const handleFilterChange = (event: FormChangeEvent<DrillDownFiltersFormValues>): void => {
        if (event.valid) {
            onFiltersChange({
                includeRepoRegexp: event.values.includeRepoRegexp,
                excludeRepoRegexp: event.values.excludeRepoRegexp,
            })
        }
    }

    const handleFilterSave = (values: DrillDownFiltersFormValues): SubmissionResult =>
        onFilterSave({
            includeRepoRegexp: values.includeRepoRegexp,
            excludeRepoRegexp: values.excludeRepoRegexp,
        })

    // By default always render filters mode
    const [step, setStep] = useState(DrillDownFiltersStep.Filters)

    if (step === DrillDownFiltersStep.Filters) {
        return (
            <DrillDownFiltersForm
                className={styles.filtersForm}
                initialFiltersValue={initialFiltersValue}
                originalFiltersValue={originalFiltersValue}
                onFiltersChange={handleFilterChange}
                onFilterSave={handleFilterSave}
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
