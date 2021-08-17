import { uniqBy } from 'lodash'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    InsightFields,
    InsightsResult,
    SubjectSettingsResult,
    SubjectSettingsVariables,
} from '../../../../graphql-operations'
import { BackendInsightFilters } from '../types'

const insightFieldsFragment = gql`
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
            }
        }
    }
`

export function fetchBackendInsights(
    insightsIds: string[],
    filters?: BackendInsightFilters
): Observable<InsightFields[]> {
    return requestGraphQL<InsightsResult>(
        gql`
            query Insights($ids: [ID!]!, $includeRepoRegex: String, $excludeRepoRegex: String) {
                insights(ids: $ids) {
                    nodes {
                        ...InsightFields
                    }
                }
            }
            ${insightFieldsFragment}
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

export function fetchLatestSubjectSettings(id: string): Observable<SubjectSettingsResult> {
    return requestGraphQL<SubjectSettingsResult, SubjectSettingsVariables>(
        gql`
            query SubjectSettings($id: ID!) {
                settingsSubject(id: $id) {
                    latestSettings {
                        id
                        contents
                    }
                }
            }
        `,
        { id }
    ).pipe(map(dataOrThrowErrors))
}
