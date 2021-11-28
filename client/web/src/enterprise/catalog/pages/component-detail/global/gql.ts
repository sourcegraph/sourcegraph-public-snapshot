import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
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
        commits(first: 7) {
            nodes {
                ...GitCommitFields
            }
        }
    }
    ${gitCommitFragment}
`

const CATALOG_COMPONENT_AUTHORS_FRAGMENT = gql`
    fragment CatalogComponentAuthorsFields on CatalogComponent {
        authors {
            person {
                ...PersonLinkFields
            }
            authoredLineCount
            authoredLineProportion
            lastCommit {
                ...GitCommitFields
            }
        }
    }
    ${gitCommitFragment}
    ${personLinkFieldsFragment}
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
        ...CatalogComponentAuthorsFields
    }
    ${CATALOG_COMPONENT_SOURCES_FRAGMENT}
    ${CATALOG_COMPONENT_CHANGES_FRAGMENT}
    ${CATALOG_COMPONENT_AUTHORS_FRAGMENT}
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
