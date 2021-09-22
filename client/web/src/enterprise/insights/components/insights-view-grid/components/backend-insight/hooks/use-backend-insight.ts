import { useMemo } from 'react'
import { LineChartContent } from 'sourcegraph'

import { gql, useQuery } from '@sourcegraph/shared/src/graphql/graphql'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { InsightFields, InsightsResult, InsightsVariables } from '../../../../../../../graphql-operations'
import { BackendInsightSeries, InsightFilters } from '../../../../../../../schema/settings.schema'
import { createViewContent } from '../../../../../core/backend/utils/create-view-content'

const INSIGHT_FIELDS_FRAGMENT = gql`
    fragment InsightFields on Insight {
        id
        title
        description
        series {
            label
            points(excludeRepoRegex: $excludeRepoRegex, includeRepoRegex: $includeRepoRegex) {
                dateTime
                value
            }
            status {
                pendingJobs
                completedJobs
                failedJobs
                backfillQueuedAt
            }
        }
    }
`

const BACKEND_INSIGHT_QUERY = gql`
    query Insights($ids: [ID!]!, $includeRepoRegex: String, $excludeRepoRegex: String) {
        insights(ids: $ids) {
            nodes {
                ...InsightFields
            }
        }
    }
    ${INSIGHT_FIELDS_FRAGMENT}
`

const createInsightView = (backendInsight: InsightFields, series: BackendInsightSeries[]): BackendInsightData => ({
    id: backendInsight.id,
    view: {
        title: backendInsight.title,
        subtitle: backendInsight.description,
        content: [createViewContent(backendInsight, series)],
        isFetchingHistoricalData: backendInsight.series.some(
            ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
        ),
    },
})

export class InsightStillProcessingError extends Error {
    constructor(message: string = 'Your insight is being processed') {
        super(message)
        this.name = 'InProcessError'
    }
}

interface UseBackendInsightProps {
    id: string
    series: BackendInsightSeries[]
    filters?: InsightFilters
}

interface UseBackendInsightResult {
    data: BackendInsightData | undefined
    loading: boolean
    error?: ErrorLike
}

interface BackendInsightData {
    id: string
    view: {
        title: string
        subtitle: string
        content: LineChartContent<any, string>[]
        isFetchingHistoricalData: boolean
    }
}

export function useBackendInsight(props: UseBackendInsightProps): UseBackendInsightResult {
    const { id, series, filters } = props

    const variables = useMemo(
        () => ({
            ids: [id],
            excludeRepoRegex: filters?.excludeRepoRegexp ?? null,
            includeRepoRegex: filters?.includeRepoRegexp ?? null,
        }),
        [filters?.excludeRepoRegexp, filters?.includeRepoRegexp, id]
    )

    const { data, loading, error } = useQuery<InsightsResult, InsightsVariables>(BACKEND_INSIGHT_QUERY, {
        variables,
        context: { concurrent: true },
        fetchPolicy: 'cache-and-network',
    })

    if (data?.insights?.nodes) {
        const backendInsights = data.insights.nodes

        if (backendInsights.length === 0) {
            return {
                data: undefined,
                loading: false,
                error: new InsightStillProcessingError(),
            }
        }

        return {
            data: createInsightView(backendInsights[0], series),
            loading,
            error,
        }
    }

    return { loading, error, data: undefined }
}
