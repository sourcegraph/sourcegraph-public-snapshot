import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const TREE_OR_COMPONENT_PAGE = gql`
    query TreeOrComponentPage($repo: ID!, $commitID: String!, $inputRevspec: String!, $path: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                commit(rev: $commitID, inputRevspec: $inputRevspec) {
                    id
                    tree(path: $path) {
                        path
                    }
                }
                primaryComponents: components(path: $path, primary: true, recursive: false) {
                    id
                    name
                }
                otherComponents: components(path: $path, primary: false, recursive: false) {
                    id
                    name
                }
            }
        }
    }
`
