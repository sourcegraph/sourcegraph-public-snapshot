import { uniqBy } from 'lodash'
import { Observable, of, throwError } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../../../../backend/graphql'
import { InsightFields, InsightsResult } from '../../../../../../graphql-operations'
import { BackendInsight, isCaptureGroupInsight } from '../../../types'
import { SearchBasedBackendFilters } from '../../../types/insight/search-insight'
import { BackendInsightData } from '../../code-insights-backend-types'
import { createLineChartContent } from '../../utils/create-line-chart-content'
import { InsightInProcessError } from '../../utils/errors'

export function getBackendInsight(insight: BackendInsight): Observable<BackendInsightData> {
    if (isCaptureGroupInsight(insight)) {
        throw new Error("Setting based API doesn't support capture group insights")
    }
    const { id, filters, series } = insight

    return fetchBackendInsights([id], filters).pipe(
        switchMap(backendInsights => {
            if (backendInsights.length === 0) {
                return throwError(new InsightInProcessError())
            }

            return of(backendInsights[0])
        }),
        map(backendInsight => ({
            id: backendInsight.id,
            view: {
                title: insight.title ?? backendInsight.title,
                subtitle: '',
                content: [createLineChartContent(backendInsight, series)],
                isFetchingHistoricalData: backendInsight.series.some(
                    ({ status: { pendingJobs, backfillQueuedAt } }) => pendingJobs > 0 || backfillQueuedAt === null
                ),
            },
        }))
    )
}

const INSIGHT_FIELDS_FRAGMENT = gql`
    fragment InsightFields on Insight {
        id
        title
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

function fetchBackendInsights(insightsIds: string[], filters?: SearchBasedBackendFilters): Observable<InsightFields[]> {
    return requestGraphQL<InsightsResult>(
        gql`
            query Insights($ids: [ID!]!, $includeRepoRegex: String, $excludeRepoRegex: String) {
                insights(ids: $ids) {
                    nodes {
                        ...InsightFields
                    }
                }
            }
            ${INSIGHT_FIELDS_FRAGMENT}
        `,
        {
            ids: insightsIds,
            excludeRepoRegex: filters?.excludeRepoRegexp ?? null,
            includeRepoRegex: filters?.includeRepoRegexp ?? null,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.insights?.nodes ?? []),
        map(data => uniqBy(data, 'id'))
    )
}
