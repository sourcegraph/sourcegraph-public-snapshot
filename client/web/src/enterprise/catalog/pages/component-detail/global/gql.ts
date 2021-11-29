import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { gitCommitFragment } from '../../../../../repo/commits/RepositoryCommitsPage'

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
            callers {
                person {
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
        }
    }
`

const CATALOG_COMPONENT_DETAIL_FRAGMENT = gql`
    fragment CatalogComponentDetailFields on CatalogComponent {
        id
        kind
        name
        system
        tags
        url
        ...CatalogComponentSourcesFields
        ...CatalogComponentChangesFields
        ...CatalogComponentAuthorsFields
        ...CatalogComponentUsageFields
    }
    ${CATALOG_COMPONENT_SOURCES_FRAGMENT}
    ${CATALOG_COMPONENT_CHANGES_FRAGMENT}
    ${CATALOG_COMPONENT_AUTHORS_FRAGMENT}
    ${CATALOG_COMPONENT_USAGE_FRAGMENT}
`

export const CATALOG_COMPONENT_BY_NAME = gql`
    query CatalogComponentByName($name: String!) {
        catalogComponent(name: $name) {
            ...CatalogComponentDetailFields
        }
    }
    ${CATALOG_COMPONENT_DETAIL_FRAGMENT}
`
