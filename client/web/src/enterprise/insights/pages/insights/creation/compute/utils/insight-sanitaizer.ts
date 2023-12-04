import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../graphql-operations'
import { getSanitizedSeries } from '../../../../../components'
import { type ComputeInsight, InsightType } from '../../../../../core'
import type { CreateComputeInsightFormFields } from '../types'

export const getSanitizedComputeInsight = (values: CreateComputeInsightFormFields): ComputeInsight => ({
    id: 'newly-created-insight',
    title: values.title,
    repositories: values.repositories,
    groupBy: values.groupBy,
    type: InsightType.Compute,
    dashboards: [],
    series: getSanitizedSeries(values.series),
    isFrozen: false,
    dashboardReferenceCount: 0,
    filters: {
        excludeRepoRegexp: '',
        includeRepoRegexp: '',
        context: '',
        seriesDisplayOptions: {
            numSamples: null,
            limit: null,
            sortOptions: {
                direction: SeriesSortDirection.DESC,
                mode: SeriesSortMode.RESULT_COUNT,
            },
        },
    },
})
