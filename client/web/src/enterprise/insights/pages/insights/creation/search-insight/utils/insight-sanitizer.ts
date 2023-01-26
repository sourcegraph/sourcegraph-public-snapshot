import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../graphql-operations'
import { getSanitizedRepositories, getSanitizedSeries } from '../../../../../components'
import { MAX_NUMBER_OF_SERIES } from '../../../../../constants'
import { MinimalSearchBasedInsightData, InsightType } from '../../../../../core'
import { CreateInsightFormFields } from '../types'

/**
 * Function converter from form shape insight to insight as it is
 * presented in user/org settings.
 */
export function getSanitizedSearchInsight(rawInsight: CreateInsightFormFields): MinimalSearchBasedInsightData {
    return {
        type: InsightType.SearchBased,
        title: rawInsight.title,
        repoQuery: rawInsight.repoMode === 'search-query' ? rawInsight.repoQuery.query : '',
        repositories: rawInsight.repoMode === 'urls-list' ? getSanitizedRepositories(rawInsight.repositories) : [],
        series: getSanitizedSeries(rawInsight.series),
        step: { [rawInsight.step]: +rawInsight.stepValue },
        dashboards: [],
        filters: {
            excludeRepoRegexp: '',
            includeRepoRegexp: '',
            context: '',
            seriesDisplayOptions: {
                limit: MAX_NUMBER_OF_SERIES,
                numSamples: null,
                sortOptions: {
                    direction: SeriesSortDirection.DESC,
                    mode: SeriesSortMode.RESULT_COUNT,
                },
            },
        },
    }
}
