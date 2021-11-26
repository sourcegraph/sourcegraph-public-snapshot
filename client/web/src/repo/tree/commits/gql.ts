import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { gitCommitFragment } from '../../commits/RepositoryCommitsPage'

export const TREE_COMMITS = gql`
    query TreeCommits($repo: ID!, $revspec: String!, $first: Int, $filePath: String, $afterDate: String) {
        node(id: $repo) {
            __typename
            ... on Repository {
                id
                commit(rev: $revspec) {
                    id
                    ancestors(first: $first, path: $filePath, after: $afterDate) {
                        nodes {
                            ...GitCommitFields
                        }
                        pageInfo {
                            hasNextPage
                        }
                    }
                }
            }
        }
    }
    ${gitCommitFragment}
`
