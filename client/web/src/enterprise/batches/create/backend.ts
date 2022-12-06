import { gql } from '@sourcegraph/http-client'

import { batchSpecExecutionFieldsFragment } from '../batch-spec/execute/backend'

export const GET_BATCH_CHANGE_TO_EDIT = gql`
    query GetBatchChangeToEdit($namespace: ID!, $name: String!) {
        batchChange(namespace: $namespace, name: $name) {
            ...EditBatchChangeFields
        }
    }

    fragment EditBatchChangeFields on BatchChange {
        __typename
        id
        url
        name
        namespace {
            __typename
            id
            ... on User {
                username
                displayName
                namespaceName
                viewerCanAdminister
                url
            }
            ... on Org {
                name
                displayName
                namespaceName
                viewerCanAdminister
                url
            }
        }
        description

        viewerCanAdminister

        currentSpec {
            id
            originalInput
            createdAt
            startedAt
            state
            applyURL
        }

        batchSpecs(first: 1) {
            nodes {
                id
                originalInput
                createdAt
                startedAt
                state
                applyURL
            }
        }

        state
    }
`

export const EXECUTE_BATCH_SPEC = gql`
    mutation ExecuteBatchSpec($batchSpec: ID!, $noCache: Boolean) {
        executeBatchSpec(batchSpec: $batchSpec, noCache: $noCache) {
            ...BatchSpecExecutionFields
        }
    }

    ${batchSpecExecutionFieldsFragment}
`

// This mutation is used to create a new batch change. It creates the batch change and an
// "empty" batch spec for it.
export const CREATE_EMPTY_BATCH_CHANGE = gql`
    mutation CreateEmptyBatchChange($namespace: ID!, $name: String!) {
        createEmptyBatchChange(namespace: $namespace, name: $name) {
            id
            url
        }
    }
`

// This mutation is used to create a new batch spec when the existing batch spec attached
// to a batch change has already been applied.
export const CREATE_BATCH_SPEC_FROM_RAW = gql`
    mutation CreateBatchSpecFromRaw($spec: String!, $namespace: ID!, $batchChange: ID!) {
        createBatchSpecFromRaw(batchSpec: $spec, namespace: $namespace, batchChange: $batchChange) {
            id
            createdAt
            workspaceResolution {
                # We fetch started at to make sure we distinguish a new workspace
                # resolution from a previous one.
                startedAt
                state
                failureMessage
            }
        }
    }
`

// This mutation is used to update the batch spec when the existing batch spec is
// unapplied.
export const REPLACE_BATCH_SPEC_INPUT = gql`
    mutation ReplaceBatchSpecInput($previousSpec: ID!, $spec: String!) {
        replaceBatchSpecInput(previousSpec: $previousSpec, batchSpec: $spec) {
            id
            createdAt
            workspaceResolution {
                # We fetch started at to make sure we distinguish a new workspace
                # resolution from a previous one.
                startedAt
                state
                failureMessage
            }
        }
    }
`

export const WORKSPACE_RESOLUTION_STATUS = gql`
    query WorkspaceResolutionStatus($batchSpec: ID!) {
        node(id: $batchSpec) {
            __typename
            ... on BatchSpec {
                workspaceResolution {
                    # We fetch started at to make sure we distinguish a new workspace
                    # resolution from a previous one.
                    startedAt
                    state
                    failureMessage
                }
            }
        }
    }
`

export const WORKSPACES = gql`
    query BatchSpecWorkspacesPreview($batchSpec: ID!, $first: Int, $after: String, $search: String) {
        node(id: $batchSpec) {
            __typename
            ... on BatchSpec {
                workspaceResolution {
                    __typename
                    workspaces(first: $first, after: $after, search: $search) {
                        __typename
                        totalCount
                        pageInfo {
                            hasNextPage
                            endCursor
                        }
                        nodes {
                            __typename
                            ... on HiddenBatchSpecWorkspace {
                                ...PreviewHiddenBatchSpecWorkspaceFields
                            }
                            ... on VisibleBatchSpecWorkspace {
                                ...PreviewVisibleBatchSpecWorkspaceFields
                            }
                        }
                    }
                }
            }
        }
    }

    fragment PreviewBatchSpecWorkspaceFields on BatchSpecWorkspace {
        __typename
        id
        ignored
        unsupported
        cachedResultFound
        stepCacheResultCount
    }

    fragment PreviewVisibleBatchSpecWorkspaceFields on VisibleBatchSpecWorkspace {
        __typename
        ...PreviewBatchSpecWorkspaceFields
        repository {
            __typename
            id
            name
            url
        }
        branch {
            __typename
            id
            displayName
            target {
                __typename
                oid
            }
            url
        }
        path
        searchResultPaths
    }

    fragment PreviewHiddenBatchSpecWorkspaceFields on HiddenBatchSpecWorkspace {
        __typename
        ...PreviewBatchSpecWorkspaceFields
    }
`

export const IMPORTING_CHANGESETS = gql`
    query BatchSpecImportingChangesets($batchSpec: ID!, $first: Int, $after: String) {
        node(id: $batchSpec) {
            __typename
            ... on BatchSpec {
                importingChangesets(first: $first, after: $after) {
                    __typename
                    totalCount
                    pageInfo {
                        hasNextPage
                        endCursor
                    }
                    nodes {
                        __typename
                        ... on VisibleChangesetSpec {
                            ...PreviewBatchSpecImportingChangesetFields
                        }
                        ... on HiddenChangesetSpec {
                            ...PreviewBatchSpecImportingHiddenChangesetFields
                        }
                    }
                }
            }
        }
    }

    fragment PreviewBatchSpecImportingChangesetFields on VisibleChangesetSpec {
        __typename
        id
        description {
            __typename
            ... on ExistingChangesetReference {
                baseRepository {
                    name
                    url
                }
                externalID
            }
        }
    }

    fragment PreviewBatchSpecImportingHiddenChangesetFields on HiddenChangesetSpec {
        __typename
        id
    }
`

export const EXECUTORS = gql`
    query CheckExecutorsAccessToken {
        areExecutorsConfigured
    }
`
