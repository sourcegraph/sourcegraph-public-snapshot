import { gql } from '@sourcegraph/http-client'

export const USER_REQUEST_QUOTAS = gql`
    query UserRequestQuotas($userID: ID!) {
        site {
            perUserCompletionsQuota
        }
        node(id: $userID) {
            __typename
            ... on User {
                completionsQuotaOverride
            }
        }
    }
`

export const SET_USER_COMPLETIONS_QUOTA = gql`
    mutation SetUserCompletionsQuota($userID: ID!, $quota: Int) {
        setUserCompletionsQuota(user: $userID, quota: $quota) {
            id
            completionsQuotaOverride
        }
    }
`
