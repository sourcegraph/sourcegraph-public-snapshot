import { type Validator, createRequiredValidator } from '@sourcegraph/wildcard'

import type { EditableDataSeries } from '../form-series'

// Group of shared Creation UI/Edit UI insight validators.

/**
 * Primarily used in any place where we edit or create insights, like creation ui page
 * or drill down insight creation flow.
 */
export const insightTitleValidator = createRequiredValidator('Title is a required field.')

/**
 * Primarily used in creation and edit insight pages and also on the landing page where
 * we have a creation UI insight sandbox demo widget.
 */
export const insightRepositoriesValidator: Validator<string[]> = value => {
    if (value !== undefined && value.length === 0) {
        return 'Repositories is a required field.'
    }

    return
}

/**
 * Validator that should be used for the time interval settings control.
 * Like Search-Based or Capture group insights time interval controls.
 */
export const insightStepValueValidator = createRequiredValidator('Please specify a step between points.')

/**
 * Custom validator for chart series. Since series has complex type
 * we can't validate this with standard validators.
 */
export const insightSeriesValidator: Validator<EditableDataSeries[]> = series => {
    if (!series || series.length === 0) {
        return 'No series defined. You must add at least one series to create a code insight.'
    }

    if (series.some(series => !series.valid)) {
        return 'Some series is invalid. Remove or edit the invalid series.'
    }

    return
}
