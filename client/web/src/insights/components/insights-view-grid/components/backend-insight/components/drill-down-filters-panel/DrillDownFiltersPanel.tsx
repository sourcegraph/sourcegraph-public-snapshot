import React, { useState } from 'react'

import {
    SearchBackendBasedInsightFiltersType,
    SearchBasedBackendFilters,
} from '../../../../../../core/types/insight/search-insight'
import { FormChangeEvent, SubmissionResult } from '../../../../../form/hooks/useForm'

import {
    DrillDownFiltersForm,
    DrillDownFiltersFormValues,
} from './components/drill-down-filters-form/DrillDownFiltersForm'
import styles from './DrillDownFiltersPanel.module.scss'
import { getDrillDownFormValues } from './utils'

enum DrillDownFiltersStep {
    Filters = 'filters',
    ViewCreation = 'view-creation',
}

export interface DrillDownFiltersPanelProps {
    initialFiltersValue: SearchBasedBackendFilters
    onFiltersChange: (filters: SearchBasedBackendFilters) => void
    onFilterSave: (filters: SearchBasedBackendFilters) => SubmissionResult
}

export const DrillDownFiltersPanel: React.FunctionComponent<DrillDownFiltersPanelProps> = props => {
    const { initialFiltersValue, onFiltersChange, onFilterSave } = props

    const handleFilterChange = (event: FormChangeEvent<DrillDownFiltersFormValues>): void => {
        if (event.valid) {
            onFiltersChange({
                type: SearchBackendBasedInsightFiltersType.Regex,
                includeRepoRegexp: event.values.includeRepoRegexp,
                excludeRepoRegexp: event.values.excludeRepoRegexp,
            })
        }
    }

    const handleFilterSave = (values: DrillDownFiltersFormValues): SubmissionResult =>
        onFilterSave({
            type: SearchBackendBasedInsightFiltersType.Regex,
            includeRepoRegexp: values.includeRepoRegexp,
            excludeRepoRegexp: values.excludeRepoRegexp,
        })

    // By default always render filters mode
    const [step] = useState(DrillDownFiltersStep.Filters)

    if (step === DrillDownFiltersStep.Filters) {
        return (
            <DrillDownFiltersForm
                className={styles.filtersForm}
                initialFiltersValue={getDrillDownFormValues(initialFiltersValue)}
                onFiltersChange={handleFilterChange}
                onFilterSave={handleFilterSave}
            />
        )
    }

    return <span>Create new insight view with filters</span>
}
