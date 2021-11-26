import { useQuery, gql } from '@sourcegraph/http-client'

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
