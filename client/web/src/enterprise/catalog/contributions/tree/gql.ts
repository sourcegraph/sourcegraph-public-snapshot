import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { COMPONENT_OWNER_FRAGMENT } from '../../components/entity-owner/gql'

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
        __typename
        id
        name
        description
        kind
        lifecycle
        labels {
            key
            values
        }
        catalogURL
        url
        ...ComponentOwnerFields
    }
    fragment OtherComponentForTreeFields on Component {
        __typename
        id
        name
        kind
        url
    }
    ${COMPONENT_OWNER_FRAGMENT}
`
