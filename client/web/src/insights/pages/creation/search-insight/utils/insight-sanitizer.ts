import { camelCase } from 'lodash'

import { DataSeries } from '../../../../core/backend/types'
import { InsightTypeSuffix, SearchBasedInsight } from '../../../../core/types'
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
 * Function converter from form shape insight to insight as it is
 * presented in user/org settings.
 */
export function getSanitizedSearchInsight(rawInsight: CreateInsightFormFields): SearchBasedInsight {
    return {
        // ID generated according to our naming insight convention
        // <Type of insight>.insight.<name of insight>
        id: `${InsightTypeSuffix.search}.${camelCase(rawInsight.title)}`,
        visibility: rawInsight.visibility,
        title: rawInsight.title,
        repositories: getSanitizedRepositories(rawInsight.repositories),
        series: getSanitizedSeries(rawInsight.series),
        step: { [rawInsight.step]: +rawInsight.stepValue },
    }
}
