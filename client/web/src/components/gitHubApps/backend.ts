import { QueryResult } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import { GitHubAppByAppIDResult, GitHubAppByAppIDVariables } from '../../graphql-operations'
import { ExternalServiceFieldsWithConfig, LIST_EXTERNAL_SERVICE_FRAGMENT } from '../externalServices/backend'

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

export const GITHUB_APPS_WITH_INSTALLATIONS_QUERY = gql`
    query GitHubAppsWithInstalls {
        gitHubApps {
            nodes {
                id
                appID
                name
                baseURL
                logo
                installations {
                    id
                    account {
                        login
                        avatarURL
                        url
                        type
                    }
                }
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
            baseURL
            name
            slug
            appURL
            baseURL
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
                externalServices(first: 100) {
                    nodes {
                        ...ListExternalServiceFields
                    }
                    totalCount
                }
            }
            webhook {
                id
            }
        }
    }
`

export const GITHUB_APP_BY_APP_ID_QUERY = gql`
    query GitHubAppByAppID($appID: Int!, $baseURL: String!) {
        gitHubAppByAppID(appID: $appID, baseURL: $baseURL) {
            id
            name
        }
    }
`

export const GITHUB_APP_CLIENT_SECRET_QUERY = gql`
    query GitHubAppClientSecret($id: ID!) {
        gitHubApp(id: $id) {
            id
            clientSecret
        }
    }
`

export const SITE_SETTINGS_QUERY = gql`
    query SiteConfigForApps {
        site {
            __typename
            id
            configuration {
                effectiveContents
            }
        }
    }
`

export const DELETE_GITHUB_APP_BY_ID_QUERY = gql`
    mutation DeleteGitHubApp($gitHubApp: ID!) {
        deleteGitHubApp(gitHubApp: $gitHubApp) {
            alwaysNil
        }
    }
`

export const useFetchGithubAppForES = (
    externalService?: ExternalServiceFieldsWithConfig
): QueryResult<GitHubAppByAppIDResult, GitHubAppByAppIDVariables> =>
    useQuery<GitHubAppByAppIDResult, GitHubAppByAppIDVariables>(GITHUB_APP_BY_APP_ID_QUERY, {
        skip: !externalService?.parsedConfig?.gitHubAppDetails,
        variables: {
            appID: externalService?.parsedConfig?.gitHubAppDetails?.appID ?? 0,
            baseURL:
                externalService?.parsedConfig?.gitHubAppDetails?.baseURL ?? externalService?.parsedConfig?.url ?? '',
        },
    })
