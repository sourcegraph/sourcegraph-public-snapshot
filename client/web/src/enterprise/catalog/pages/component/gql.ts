import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { COMPONENT_OWNER_FRAGMENT } from '../../components/entity-owner/gql'

const COMPONENT_WHO_KNOWS_FRAGMENT = gql`
    fragment ComponentWhoKnowsFields on Component {
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

const COMPONENT_CODE_OWNERS_FRAGMENT = gql`
    fragment ComponentCodeOwnersFields on Component {
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

const COMPONENT_STATUS_FRAGMENT = gql`
    fragment ComponentStatusFields on Component {
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

const COMPONENT_DOCUMENTATION_FRAGMENT = gql`
    fragment ComponentDocumentationFields on Component {
        readme {
            name
            richHTML
            url
        }
    }
`

const COMPONENT_SOURCES_FRAGMENT = gql`
    fragment ComponentSourcesFields on Component {
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

const COMPONENT_CHANGES_FRAGMENT = gql`
    fragment ComponentChangesFields on Component {
        commits(first: 10) {
            nodes {
                ...GitCommitFields
            }
        }
    }
    ${gitCommitFragment}
`

const COMPONENT_AUTHORS_FRAGMENT = gql`
    fragment ComponentAuthorsFields on Component {
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

const COMPONENT_USAGE_FRAGMENT = gql`
    fragment ComponentUsageFields on Component {
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

const COMPONENT_API_FRAGMENT = gql`
    fragment ComponentAPIFields on Component {
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

export const COMPONENT_DETAIL_FRAGMENT = gql`
    fragment ComponentStateDetailFields on Component {
        __typename
        id
        name
        kind
        description
        lifecycle
        url
        ...ComponentOwnerFields
        ...ComponentStatusFields
        ...ComponentCodeOwnersFields
        ...ComponentWhoKnowsFields
        ...ComponentDocumentationFields
        ...ComponentSourcesFields
        ...ComponentChangesFields
        ...ComponentAuthorsFields
        ...ComponentUsageFields
        ...ComponentAPIFields
    }
    ${COMPONENT_WHO_KNOWS_FRAGMENT}
    ${COMPONENT_OWNER_FRAGMENT}
    ${COMPONENT_STATUS_FRAGMENT}
    ${COMPONENT_CODE_OWNERS_FRAGMENT}
    ${COMPONENT_DOCUMENTATION_FRAGMENT}
    ${COMPONENT_SOURCES_FRAGMENT}
    ${COMPONENT_CHANGES_FRAGMENT}
    ${COMPONENT_AUTHORS_FRAGMENT}
    ${COMPONENT_USAGE_FRAGMENT}
    ${COMPONENT_API_FRAGMENT}
`

export const COMPONENT_BY_NAME = gql`
    query ComponentByName($name: String!) {
        component(name: $name) {
            ...ComponentStateDetailFields
        }
    }
    ${COMPONENT_DETAIL_FRAGMENT}
`
