import { gql } from '@sourcegraph/http-client'

export const USER_REQUEST_QUOTAS = gql`
    query UserRequestQuotas($userID: ID!) {
        site {
            perUserCompletionsQuota
            perUserCodeCompletionsQuota
        }
        node(id: $userID) {
            __typename
            ... on User {
                completionsQuotaOverride
                codeCompletionsQuotaOverride
                completionsQuotaOverrideNote
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

export const SET_USER_CODE_COMPLETIONS_QUOTA = gql`
    mutation SetUserCodeCompletionsQuota($userID: ID!, $quota: Int) {
        setUserCodeCompletionsQuota(user: $userID, quota: $quota) {
            id
            codeCompletionsQuotaOverride
        }
    }
`

export const SET_USER_COMPLETIONS_QUOTA_NOTE = gql`
    mutation SetUserCompletionsQuotaNote($userID: ID!, $note: String!) {
        setUserCompletionsQuotaNote(user: $userID, note: $note) {
            id
            completionsQuotaOverrideNote
        }
    }
`
