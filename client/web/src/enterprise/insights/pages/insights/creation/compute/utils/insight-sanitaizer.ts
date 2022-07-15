import { getSanitizedRepositories, getSanitizedSeries } from '../../../../../components'
import { ComputeInsight, InsightExecutionType, InsightType } from '../../../../../core'
import { CreateComputeInsightFormFields } from '../types'

export const getSanitizedComputeInsight = (values: CreateComputeInsightFormFields): ComputeInsight => ({
    id: 'newly-created-insight',
    title: values.title,
    repositories: getSanitizedRepositories(values.repositories),
    groupBy: values.groupBy,
    type: InsightType.Compute,
    executionType: InsightExecutionType.Backend,
    dashboards: [],
    series: getSanitizedSeries(values.series),
    isFrozen: false,
    dashboardReferenceCount: 0,
    filters: {
        excludeRepoRegexp: '',
        includeRepoRegexp: '',
        context: '',
    },
})
