import type { Duration } from 'date-fns'

import { DataSeries } from '../../../../core/backend/types'
import { CreateInsightFormFields } from '../types'

export function getSanitizedRepositories(rawRepositories: string): string[] {
    return rawRepositories.trim().split(/\s*,\s*/)
}

export function getSanitizedSeries(rawSeries: DataSeries[]): DataSeries[] {
    return rawSeries.map(line => ({
        ...line,
        // Query field is a reg exp field for code insight query setting
        // Native html input element adds escape symbols by itself
        // to prevent this behavior below we replace double escaping
        // with just one series of escape characters e.g. - //
        query: line.query.replace(/\\\\/g, '\\'),
    }))
}

/**
 * Insight as it is presented in user/org settings.
 */
export interface SanitizedInsight {
    title: string
    repositories: string[]
    series: DataSeries[]
    step: Duration
}

/**
 * Function converter from form shape insight to insight as it is
 * presented in user/org settings.
 */
export function getSanitizedInsight(rawInsight: CreateInsightFormFields): SanitizedInsight {
    return {
        title: rawInsight.title,
        repositories: getSanitizedRepositories(rawInsight.repositories),
        series: getSanitizedSeries(rawInsight.series),
        step: { [rawInsight.step]: +rawInsight.stepValue },
    }
}
