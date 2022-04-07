import { gql } from '@apollo/client'

export const GET_INSIGHTS_DASHBOARD_OWNERS_GQL = gql`
    query InsightSubjects {
        currentUser {
            id
            organizations {
                nodes {
                    id
                    name
                    displayName
                }
            }
        }
        site {
            id
        }
    }
`
