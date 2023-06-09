import { gql, useQuery } from '@sourcegraph/http-client'

import { gitCommitFragment } from '../commits/RepositoryCommitsPage'

export const COMMIT_QUERY = gql`
    query RepositoryCommit($repo: ID!, $revspec: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                sourceType
                commit(rev: $revspec) {
                    __typename # Necessary for error handling to check if commit exists
                    ...GitCommitFields
                }
            }
        }
    }
    ${gitCommitFragment}
`

export const changelistFragment = gql`
    fragment PerforceChangelistFields on PerforceChangelist {
        cid
        canonicalURL
        commit {
            ...GitCommitFields
            __typename
        }
    }
    ${gitCommitFragment}
`

export const CHANGELIST_QUERY = gql`
    query RepositoryChangelist($repo: ID!, $changelistID: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                sourceType
                changelist(cid: $changelistID) {
                    __typename # Necessary for error handling to check if commit exists
                    ...PerforceChangelistFields
                }
            }
        }
    }
    ${changelistFragment}
`
