import { gql } from '@sourcegraph/http-client'

import { LIST_EXTERNAL_SERVICE_FRAGMENT } from '../externalServices/backend'

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

export const GITHUB_APP_BY_ID_QUERY = gql`
    ${LIST_EXTERNAL_SERVICE_FRAGMENT}
    query GitHubAppByID($id: ID!) {
        gitHubApp(id: $id) {
            id
            appID
            name
            slug
            appURL
            clientID
            logo
            createdAt
            updatedAt
            installations {
                id
                url
                account {
                    login
                    avatarURL
                    url
                    type
                }
            }
            externalServices(first: 100) {
                nodes {
                    ...ListExternalServiceFields
                }
                totalCount
            }
        }
    }
`
