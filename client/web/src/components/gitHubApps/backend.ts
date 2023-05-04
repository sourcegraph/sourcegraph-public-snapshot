import { gql } from '@sourcegraph/http-client'

export const GITHUB_APPS_QUERY = gql`
    query GitHubApps {
        gitHubApps {
            nodes {
                id
                appID
                name
                slug
                appURL
                clientID
                logo
                createdAt
                updatedAt
            }
        }
    }
`
