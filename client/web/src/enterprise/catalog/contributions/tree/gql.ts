import { gql } from '@sourcegraph/http-client'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { COMPONENT_OWNER_LINK_FRAGMENT } from '../../components/component-owner-link/ComponentOwnerLink'
import { COMPONENT_LABELS_FRAGMENT, COMPONENT_TAGS_FRAGMENT } from '../../pages/component/gql'

const SOURCE_LOCATION_SET_FILES_FRAGMENT = gql`
    fragment SourceLocationSetFilesFields on SourceLocationSet {
        __typename
        ... on GitTree {
            repository {
                id
                name
                url
            }
            path
            ...SourceLocationSetGitTreeFilesFields
        }
        ... on Component {
            sourceLocations {
                repositoryName
                repository {
                    id
                    name
                    url
                }
                path
                treeEntry {
                    __typename
                    ... on GitBlob {
                        commit {
                            oid
                        }
                        path
                        name
                        isDirectory
                        url
                    }
                    ... on GitTree {
                        ...SourceLocationSetGitTreeFilesFields
                    }
                }
            }
        }
    }
    fragment SourceLocationSetGitTreeFilesFields on GitTree {
        commit {
            oid
        }
        entries(recursive: true) {
            path
            name
            isDirectory
            url
        }
    }
`

export const SOURCE_LOCATION_SET_README_FRAGMENT = gql`
    fragment SourceLocationSetReadmeFields on SourceLocationSet {
        readme {
            name
            richHTML
            url
        }
    }
`

const SOURCE_LOCATION_SET_LAST_COMMIT_FRAGMENT = gql`
    fragment SourceLocationSetLastCommitFields on SourceLocationSet {
        commits(first: 1) {
            nodes {
                ...GitCommitFields
            }
        }
    }
    ${gitCommitFragment}
`

// TODO(sqs): dont fetch all
const SOURCE_LOCATION_SET_CODE_OWNERS_FRAGMENT = gql`
    fragment SourceLocationSetCodeOwnersFields on SourceLocationSet {
        codeOwners {
            edges {
                node {
                    ...PersonLinkFields
                    avatarURL
                }
                fileCount
                fileProportion
            }
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
    ${personLinkFieldsFragment}
`

const SOURCE_LOCATION_SET_CONTRIBUTORS_FRAGMENT = gql`
    fragment SourceLocationSetContributorsFields on SourceLocationSet {
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
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
    ${personLinkFieldsFragment}
`

export const TREE_OR_COMPONENT_SOURCE_LOCATION_SET_FRAGMENT = gql`
    fragment TreeOrComponentSourceLocationSetFields on SourceLocationSet {
        id
        ...SourceLocationSetFilesFields
        ...SourceLocationSetReadmeFields
        ...SourceLocationSetLastCommitFields
        ...SourceLocationSetCodeOwnersFields
        ...SourceLocationSetContributorsFields
        branches(first: 0, interactive: false) {
            totalCount
        }
        usage {
            __typename
        }
    }
    ${SOURCE_LOCATION_SET_FILES_FRAGMENT}
    ${SOURCE_LOCATION_SET_README_FRAGMENT}
    ${SOURCE_LOCATION_SET_LAST_COMMIT_FRAGMENT}
    ${SOURCE_LOCATION_SET_CODE_OWNERS_FRAGMENT}
    ${SOURCE_LOCATION_SET_CONTRIBUTORS_FRAGMENT}
`

export const TREE_OR_COMPONENT_PAGE = gql`
    query TreeOrComponentPage($repo: ID!, $commitID: String!, $inputRevspec: String!, $path: String!) {
        node(id: $repo) {
            __typename
            ... on Repository {
                __typename
                id
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
        name
        isRoot
        url
        ...TreeOrComponentSourceLocationSetFields
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
        ...TreeOrComponentSourceLocationSetFields
        ...ComponentOwnerLinkFields
        ...ComponentTagsFields
        ...ComponentLabelsFields
    }
    fragment OtherComponentForTreeFields on Component {
        __typename
        id
        name
        kind
        url
    }

    ${TREE_OR_COMPONENT_SOURCE_LOCATION_SET_FRAGMENT}
    ${COMPONENT_OWNER_LINK_FRAGMENT}
    ${COMPONENT_TAGS_FRAGMENT}
    ${COMPONENT_LABELS_FRAGMENT}
`
