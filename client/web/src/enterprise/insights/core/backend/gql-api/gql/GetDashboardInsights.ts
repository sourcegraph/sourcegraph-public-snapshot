import { gql } from '@apollo/client'

import { INSIGHT_VIEW_FRAGMENT } from './GetInsights'

export const GET_DASHBOARD_INSIGHTS_GQL = gql`
    query GetDashboardInsights($id: ID) {
        insightsDashboards(id: $id) {
            nodes {
                id
                views {
                    nodes {
                        ...InsightViewNode
                    }
                }
            }
        }
    }
    ${INSIGHT_VIEW_FRAGMENT}
`
