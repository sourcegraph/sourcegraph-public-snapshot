import { gql } from '@sourcegraph/http-client'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { COMPONENT_TAG_FRAGMENT } from '../../components/component-tag/ComponentTag'
import { COMPONENT_OWNER_FRAGMENT } from '../../pages/component/meta/ComponentOwnerSidebarItem'
import { SOURCE_LOCATION_SET_README_FRAGMENT } from '../../pages/component/readme/ComponentReadme'

import { SOURCE_SET_DESCENDENT_COMPONENTS_FRAGMENT } from './SourceSetDescendentComponents'

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
                isPrimary
                repositoryName
                repository {
                    id
                    name
                    url
                }
                path
                treeEntry {
                    __typename
                    url
                    ... on GitBlob {
                        commit {
                            oid
                        }
                        path
                        name
                        isDirectory
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

        ...SourceSetDescendentComponentsFields
        ...SourceLocationSetFilesFields
        ...SourceLocationSetReadmeFields
        ...SourceLocationSetCodeOwnersFields
        ...SourceLocationSetContributorsFields

        branches(first: 0, interactive: false) {
            totalCount
        }

        commitsForLastCommit: commits(first: 1) {
            nodes {
                ...GitCommitFields
            }
        }

        usage {
            __typename
        }
    }
    ${SOURCE_SET_DESCENDENT_COMPONENTS_FRAGMENT}
    ${SOURCE_LOCATION_SET_FILES_FRAGMENT}
    ${SOURCE_LOCATION_SET_README_FRAGMENT}
    ${SOURCE_LOCATION_SET_CODE_OWNERS_FRAGMENT}
    ${SOURCE_LOCATION_SET_CONTRIBUTORS_FRAGMENT}
    ${gitCommitFragment}
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
        catalogURL
        url
        ...TreeOrComponentSourceLocationSetFields
        ...ComponentOwnerFields
        tags {
            ...ComponentTagFields
        }
    }

    ${TREE_OR_COMPONENT_SOURCE_LOCATION_SET_FRAGMENT}
    ${COMPONENT_OWNER_FRAGMENT}
    ${COMPONENT_TAG_FRAGMENT}
`
