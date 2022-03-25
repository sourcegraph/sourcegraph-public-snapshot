import { gql } from '@apollo/client'

const INSIGHTS_DASHBOARD_FRAGMENT = gql`
    fragment InsightsDashboardNode on InsightsDashboard {
        id
        title
        views {
            nodes {
                id
            }
        }
        grants {
            users
            organizations
            global
        }
    }
`

const INSIGHTS_DASHBOARD_CURRENT_USER_FRAGMENT = gql`
    fragment InsightsDashboardCurrentUser on User {
        id
        organizations {
            nodes {
                id
                displayName
            }
        }
    }
`

export const GET_INSIGHTS_DASHBOARDS_GQL = gql`
    query InsightsDashboards($id: ID) {
        currentUser {
            ...InsightsDashboardCurrentUser
        }
        insightsDashboards(id: $id) {
            nodes {
                ...InsightsDashboardNode
            }
        }
    }
    ${INSIGHTS_DASHBOARD_FRAGMENT}
    ${INSIGHTS_DASHBOARD_CURRENT_USER_FRAGMENT}
`
