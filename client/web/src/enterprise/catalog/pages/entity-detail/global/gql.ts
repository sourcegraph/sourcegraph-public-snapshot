import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { gitCommitFragment } from '../../../../../repo/commits/RepositoryCommitsPage'
import { CATALOG_ENTITY_OWNER_FRAGMENT } from '../../../components/entity-owner/gql'

const CATALOG_ENTITY_WHO_KNOWS_FRAGMENT = gql`
    fragment CatalogEntityWhoKnowsFields on CatalogEntity {
        whoKnows {
            node {
                ...PersonLinkFields
                avatarURL
            }
            reasons
            score
        }
    }
    ${personLinkFieldsFragment}
`

const CATALOG_ENTITY_CODE_OWNERS_FRAGMENT = gql`
    fragment CatalogEntityCodeOwnersFields on CatalogEntity {
        codeOwners {
            node {
                ...PersonLinkFields
                avatarURL
            }
            fileCount
            fileProportion
        }
    }
    ${personLinkFieldsFragment}
`

const CATALOG_ENTITY_STATUS_FRAGMENT = gql`
    fragment CatalogEntityStatusFields on CatalogEntity {
        status {
            id
            contexts {
                id
                name
                state
                title
                description
                targetURL
            }
        }
    }
`

const CATALOG_COMPONENT_DOCUMENTATION_FRAGMENT = gql`
    fragment CatalogComponentDocumentationFields on CatalogComponent {
        readme {
            name
            richHTML
            url
        }
    }
`

const CATALOG_COMPONENT_SOURCES_FRAGMENT = gql`
    fragment CatalogComponentSourcesFields on CatalogComponent {
        sourceLocations {
            path
            isDirectory
            url
            ... on GitTree {
                repository {
                    name
                    url
                }
                files(recursive: true) {
                    path
                    name
                    isDirectory
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
        commits(first: 10) {
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
                avatarURL
            }
            authoredLineCount
            authoredLineProportion
            lastCommit {
                author {
                    date
                }
            }
        }
    }
    ${gitCommitFragment}
    ${personLinkFieldsFragment}
`

const CATALOG_COMPONENT_USAGE_FRAGMENT = gql`
    fragment CatalogComponentUsageFields on CatalogComponent {
        usage {
            locations {
                nodes {
                    range {
                        start {
                            line
                            character
                        }
                        end {
                            line
                            character
                        }
                    }
                    resource {
                        path
                        commit {
                            oid
                        }
                        repository {
                            name
                        }
                    }
                }
            }
            people {
                node {
                    ...PersonLinkFields
                    avatarURL
                }
                authoredLineCount
                lastCommit {
                    author {
                        date
                    }
                }
            }
            components {
                node {
                    id
                    name
                    kind
                    url
                }
            }
        }
    }
`

const CATALOG_COMPONENT_API_FRAGMENT = gql`
    fragment CatalogComponentAPIFields on CatalogComponent {
        api {
            symbols {
                __typename
                nodes {
                    ...SymbolFields
                }
                pageInfo {
                    hasNextPage
                }
            }
            schema {
                __typename
                path
                url
                ... on GitBlob {
                    commit {
                        oid
                    }
                    repository {
                        name
                        url
                    }
                }
            }
        }
    }

    fragment SymbolFields on Symbol {
        __typename
        name
        containerName
        kind
        language
        fileLocal
        location {
            resource {
                path
            }
            range {
                start {
                    line
                    character
                }
                end {
                    line
                    character
                }
            }
        }
        url
    }
`

export const CATALOG_ENTITY_DETAIL_FRAGMENT = gql`
    fragment CatalogEntityDetailFields on CatalogEntity {
        __typename
        id
        type
        name
        description
        url
        ...CatalogEntityOwnerFields
        ...CatalogEntityStatusFields
        ...CatalogEntityCodeOwnersFields
        ...CatalogEntityWhoKnowsFields
        ... on CatalogComponent {
            kind
            lifecycle
            ...CatalogComponentDocumentationFields
            ...CatalogComponentSourcesFields
            ...CatalogComponentChangesFields
            ...CatalogComponentAuthorsFields
            ...CatalogComponentUsageFields
            ...CatalogComponentAPIFields
        }
    }
    ${CATALOG_ENTITY_WHO_KNOWS_FRAGMENT}
    ${CATALOG_ENTITY_OWNER_FRAGMENT}
    ${CATALOG_ENTITY_STATUS_FRAGMENT}
    ${CATALOG_ENTITY_CODE_OWNERS_FRAGMENT}
    ${CATALOG_COMPONENT_DOCUMENTATION_FRAGMENT}
    ${CATALOG_COMPONENT_SOURCES_FRAGMENT}
    ${CATALOG_COMPONENT_CHANGES_FRAGMENT}
    ${CATALOG_COMPONENT_AUTHORS_FRAGMENT}
    ${CATALOG_COMPONENT_USAGE_FRAGMENT}
    ${CATALOG_COMPONENT_API_FRAGMENT}
`

export const CATALOG_ENTITY_BY_NAME = gql`
    query CatalogEntityByName($type: CatalogEntityType!, $name: String!) {
        catalogEntity(type: $type, name: $name) {
            ...CatalogEntityDetailFields
        }
    }
    ${CATALOG_ENTITY_DETAIL_FRAGMENT}
`
