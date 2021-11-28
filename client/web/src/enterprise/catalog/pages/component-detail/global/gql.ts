import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { gitCommitFragment } from '../../../../../repo/commits/RepositoryCommitsPage'

const CATALOG_COMPONENT_SOURCES_FRAGMENT = gql`
    fragment CatalogComponentSourcesFields on CatalogComponent {
        sourceLocation {
            path
            isDirectory
            canonicalURL
            ... on GitTree {
                repository {
                    name
                    url
                }
            }
            ... on GitBlob {
                repository {
                    name
                    url
                }
            }
        }
    }
    ${gitCommitFragment}
`

const CATALOG_COMPONENT_CHANGES_FRAGMENT = gql`
    fragment CatalogComponentChangesFields on CatalogComponent {
        editCommits(first: 7) {
            nodes {
                ...GitCommitFields
            }
        }
    }
    ${gitCommitFragment}
`

const CATALOG_COMPONENT_DETAIL_FRAGMENT = gql`
    fragment CatalogComponentDetailFields on CatalogComponent {
        id
        kind
        name
        system
        tags
        ...CatalogComponentSourcesFields
        ...CatalogComponentChangesFields
    }
    ${CATALOG_COMPONENT_SOURCES_FRAGMENT}
    ${CATALOG_COMPONENT_CHANGES_FRAGMENT}
`

export const CATALOG_COMPONENT_BY_ID = gql`
    query CatalogComponentByID($id: ID!) {
        node(id: $id) {
            __typename
            ... on CatalogComponent {
                ...CatalogComponentDetailFields
            }
        }
    }
    ${CATALOG_COMPONENT_DETAIL_FRAGMENT}
`
