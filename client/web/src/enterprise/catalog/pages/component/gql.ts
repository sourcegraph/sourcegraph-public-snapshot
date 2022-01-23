import { gql } from '@sourcegraph/http-client'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { COMPONENT_OWNER_LINK_FRAGMENT } from '../../components/component-owner-link/ComponentOwnerLink'

export const COMPONENT_LABELS_FRAGMENT = gql`
    fragment ComponentLabelsFields on Component {
        labels {
            key
            values
        }
    }
`

export const COMPONENT_CODE_OWNERS_FRAGMENT = gql`
    fragment ComponentCodeOwnersFields on Component {
        codeOwners {
            edges {
                node {
                    ...PersonLinkFields
                    avatarURL
                }
                fileCount
                fileProportion
            }
        }
    }
    ${personLinkFieldsFragment}
`

export const COMPONENT_DOCUMENTATION_FRAGMENT = gql`
    fragment ComponentDocumentationFields on Component {
        readme {
            name
            richHTML
            url
        }
    }
`

export const COMPONENT_SOURCE_LOCATIONS_FRAGMENT = gql`
    fragment ComponentSourceLocationsFields on Component {
        sourceLocations {
            repositoryName
            repository {
                name
                url
            }
            path
            treeEntry {
                __typename
                isDirectory
                url
                ... on GitTree {
                    entries(recursive: true) {
                        path
                        name
                        isDirectory
                        url
                    }
                    commit {
                        oid
                    }
                }
                ... on GitBlob {
                    commit {
                        oid
                    }
                }
            }
            isPrimary
        }
    }
`

export const COMPONENT_COMMITS_FRAGMENT = gql`
    fragment ComponentCommitsFields on Component {
        commits(first: 10) {
            nodes {
                ...GitCommitFields
            }
        }
    }
    ${gitCommitFragment}
`

export const COMPONENT_AUTHORS_FRAGMENT = gql`
    fragment ComponentAuthorsFields on Component {
        contributors {
            edges {
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
    }
    ${personLinkFieldsFragment}
`

const COMPONENT_USAGE_LOCATIONS_COMPONENTS_FRAGMENT = gql`
    fragment ComponentUsageLocationsComponentsFields on Component {
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
export const COMPONENT_USAGE_PEOPLE_FRAGMENT = gql`
    fragment ComponentUsagePeopleFields on Component {
        usage {
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
        }
    }
`

// TODO(sqs): remove
export const COMPONENT_DETAIL_FRAGMENT = gql`
    fragment ComponentDetailFields on Component {
        __typename
        id
        name
        kind
        description
        lifecycle
        url
        catalogURL
        ...ComponentLabelsFields
        ...ComponentOwnerLinkFields
        ...ComponentCodeOwnersFields
        ...ComponentDocumentationFields
        ...ComponentSourceLocationsFields
        ...ComponentCommitsFields
        ...ComponentAuthorsFields
        ...ComponentUsageLocationsComponentsFields
        ...ComponentUsagePeopleFields
    }
    ${COMPONENT_LABELS_FRAGMENT}
    ${COMPONENT_OWNER_LINK_FRAGMENT}
    ${COMPONENT_CODE_OWNERS_FRAGMENT}
    ${COMPONENT_DOCUMENTATION_FRAGMENT}
    ${COMPONENT_SOURCE_LOCATIONS_FRAGMENT}
    ${COMPONENT_COMMITS_FRAGMENT}
    ${COMPONENT_AUTHORS_FRAGMENT}
    ${COMPONENT_USAGE_LOCATIONS_COMPONENTS_FRAGMENT}
    ${COMPONENT_USAGE_PEOPLE_FRAGMENT}
`
