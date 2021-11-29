import { gql } from '@apollo/client'

export const GET_INSIGHTS_DASHBOARDS_GQL = gql`
    query InsightsDashboards($id: ID) {
        insightsDashboards(id: $id) {
            nodes {
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
        }
    }
`
