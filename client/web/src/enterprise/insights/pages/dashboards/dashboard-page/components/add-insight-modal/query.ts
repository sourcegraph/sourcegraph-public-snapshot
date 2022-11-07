import { gql } from '@apollo/client'

export const GET_ACCESSIBLE_INSIGHTS_LIST = gql`
    fragment AccessibleInsight on InsightView {
        id
        presentation {
            __typename
            ... on LineChartInsightViewPresentation {
                title
            }
            ... on PieChartInsightViewPresentation {
                title
            }
        }
    }

    query GetDashboardAccessibleInsights($id: ID!) {
        dashboardInsightsIds: insightsDashboards(id: $id) {
            nodes {
                views {
                    nodes {
                        id
                    }
                }
            }
        }

        accessibleInsights: insightViews {
            nodes {
                ...AccessibleInsight
            }
        }
    }
`
