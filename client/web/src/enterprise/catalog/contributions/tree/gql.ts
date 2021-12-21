import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const TREE_OR_COMPONENT_PAGE = gql`
    query TreeOrComponentPage($repo: ID!, $commitID: String!, $inputRevspec: String!, $path: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                ...RepositoryForTreeFields
                commit(rev: $commitID, inputRevspec: $inputRevspec) {
                    id
                    tree(path: $path) {
                        ...TreeEntryForTreeFields
                    }
                }
                primaryComponents: components(path: $path, primary: true, recursive: false) {
                    ...PrimaryComponentForTreeFields
                }
                otherComponents: components(path: $path, primary: false, recursive: false) {
                    ...OtherComponentForTreeFields
                }
            }
        }
    }

    fragment RepositoryForTreeFields on Repository {
        id
        name
        description
    }
    fragment TreeEntryForTreeFields on GitTree {
        path
        isRoot
    }
    fragment PrimaryComponentForTreeFields on Component {
        id
        name
        kind
    }
    fragment OtherComponentForTreeFields on Component {
        id
        name
        kind
    }
`
