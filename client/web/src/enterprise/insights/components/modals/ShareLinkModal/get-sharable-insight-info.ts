import { gql } from '@apollo/client'

import { GetSharableInsightInfoResult } from '../../../../../graphql-operations'

export const GET_SHARABLE_INSIGHT_INFO_GQL = gql`
    query GetSharableInsightInfo($id: ID!) {
        insightViews(id: $id) {
            nodes {
                dashboards {
                    nodes {
                        id
                    }
                }
            }
        }
    }
`

/**
 * Decodes response from the {@link GET_SHARABLE_INSIGHT_INFO_GQL} and returns
 * parsed array of dashboard ids.
 */
export function decodeDashboardIds(repsonse: GetSharableInsightInfoResult): string[] {
    if (repsonse.insightViews.nodes.length === 0) {
        throw new Error('Insight was not found')
    }

    // We get insight by id some we expect exactly one entity from the backend
    const insight = repsonse.insightViews.nodes[0]

    if (!insight.dashboards) {
        throw new Error('Insight is not included in any dashboard')
    }

    return insight.dashboards.nodes.map(dashboard => dashboard.id)
}
