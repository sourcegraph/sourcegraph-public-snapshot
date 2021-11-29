import { gql } from '@apollo/client'

export const GET_INSIGHTS_SUBJECTS_GQL = gql`
    query InsightSubjects {
        currentUser {
            __typename
            id
            displayName
            username
            viewerCanAdminister

            organizations {
                nodes {
                    __typename
                    id
                    name
                    displayName
                    viewerCanAdminister
                }
            }
        }
        site {
            __typename
            id
            allowSiteSettingsEdits
            viewerCanAdminister
        }
    }
`
