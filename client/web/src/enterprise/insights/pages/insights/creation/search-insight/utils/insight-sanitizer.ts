import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../graphql-operations'
import { getSanitizedRepositories, getSanitizedSeries } from '../../../../../components'
import { MinimalSearchBasedInsightData, InsightExecutionType, InsightType } from '../../../../../core'
import { MAX_NUMBER_OF_SERIES } from '../../../../../core/backend/gql-backend/methods/get-backend-insight-data/deserializators'
import { CreateInsightFormFields } from '../types'

/**
 * Function converter from form shape insight to insight as it is
 * presented in user/org settings.
 */
export function getSanitizedSearchInsight(rawInsight: CreateInsightFormFields): MinimalSearchBasedInsightData {
    if (rawInsight.allRepos) {
        return {
            executionType: InsightExecutionType.Backend,
            repositories: [],
            type: InsightType.SearchBased,
            title: rawInsight.title,
            series: getSanitizedSeries(rawInsight.series),
            step: { [rawInsight.step]: +rawInsight.stepValue },
            dashboards: [],
            filters: {
                excludeRepoRegexp: '',
                includeRepoRegexp: '',
                context: '',
                seriesDisplayOptions: {
                    limit: `${MAX_NUMBER_OF_SERIES}`,
                    sortOptions: {
                        direction: SeriesSortDirection.DESC,
                        mode: SeriesSortMode.RESULT_COUNT,
                    },
                },
            },
            seriesCount: 0,
        }
    }

    return {
        executionType: InsightExecutionType.Backend,
        type: InsightType.SearchBased,
        title: rawInsight.title,
        repositories: getSanitizedRepositories(rawInsight.repositories),
        series: getSanitizedSeries(rawInsight.series),
        step: { [rawInsight.step]: +rawInsight.stepValue },
        dashboards: [],
        filters: {
            excludeRepoRegexp: '',
            includeRepoRegexp: '',
            context: '',
            seriesDisplayOptions: {
                limit: `${MAX_NUMBER_OF_SERIES}`,
                sortOptions: {
                    direction: SeriesSortDirection.DESC,
                    mode: SeriesSortMode.RESULT_COUNT,
                },
            },
        },
        seriesCount: 0,
    }
}
