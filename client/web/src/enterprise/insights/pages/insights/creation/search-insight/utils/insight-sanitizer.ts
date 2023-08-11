import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../graphql-operations'
import { getSanitizedSeries } from '../../../../../components'
import { type MinimalSearchBasedInsightData, InsightType } from '../../../../../core'
import type { CreateInsightFormFields } from '../types'

/**
 * Function converter from form shape insight to insight as it is
 * presented in user/org settings.
 */
export function getSanitizedSearchInsight(rawInsight: CreateInsightFormFields): MinimalSearchBasedInsightData {
    return {
        type: InsightType.SearchBased,
        title: rawInsight.title,
        repoQuery: rawInsight.repoMode === 'search-query' ? rawInsight.repoQuery.query : '',
        repositories: rawInsight.repoMode === 'urls-list' ? rawInsight.repositories : [],
        series: getSanitizedSeries(rawInsight.series),
        step: { [rawInsight.step]: +rawInsight.stepValue },
        dashboards: [],
        filters: {
            excludeRepoRegexp: '',
            includeRepoRegexp: '',
            context: '',
            seriesDisplayOptions: {
                limit: null,
                numSamples: null,
                sortOptions: {
                    direction: SeriesSortDirection.DESC,
                    mode: SeriesSortMode.RESULT_COUNT,
                },
            },
        },
    }
}
