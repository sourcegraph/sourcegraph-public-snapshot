import { gql } from '@apollo/client'

export const REMOVE_INSIGHT_FROM_DASHBOARD_GQL = gql`
    mutation RemoveInsightViewFromDashboard($insightId: ID!, $dashboardId: ID!) {
        removeInsightViewFromDashboard(input: { insightViewId: $insightId, dashboardId: $dashboardId }) {
            dashboard {
                id
            }
        }
    }
`
