import React, { useState } from 'react';

import { FormChangeEvent } from '../../../../../form/hooks/useForm';

import { DrillDownFiltersForm } from './components/drill-down-filters-form/DrillDownFiltersForm';
import styles from './DrillDownFiltersPanel.module.scss'
import { DrillDownFilters } from './types';

enum DrillDownFiltersStep {
    Filters = 'filters',
    ViewCreation = 'view-creation'
}

export interface DrillDownFiltersPanelProps {
    initialFiltersValue?: DrillDownFilters
    onFiltersChange?: (filters: FormChangeEvent<DrillDownFilters>) => void
    onClose?: () => void
}

export const DrillDownFiltersPanel: React.FunctionComponent<DrillDownFiltersPanelProps> = props => {
    const { initialFiltersValue, onFiltersChange, onClose } = props;

    // By default always render filters mode
    const [step, setStep] = useState(DrillDownFiltersStep.Filters)

    if (step === DrillDownFiltersStep.Filters) {
        return (
            <DrillDownFiltersForm
                className={styles.filtersForm}
                initialFiltersValue={initialFiltersValue}
                onFiltersChange={onFiltersChange}/>
        )
    }

    return (
        <span>Create new insight view with filters</span>
    )
}
