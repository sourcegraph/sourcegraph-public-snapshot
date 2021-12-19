import { gql } from '@sourcegraph/http-client'

import { gitCommitFragment } from '../../commits/RepositoryCommitsPage'

export const TREE_COMMITS = gql`
    query TreeCommits(
        $repo: ID!
        $revspec: String!
        $first: Int
        $filePath: String
        $after: String
        $afterCursor: String
    ) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $revspec) {
                    ancestors(first: $first, path: $filePath, after: $after, afterCursor: $afterCursor) {
                        nodes {
                            ...GitCommitFields
                        }
                        pageInfo {
                            hasNextPage
                            endCursor
                        }
                    }
                }
            }
        }
    }
    ${gitCommitFragment}
`
