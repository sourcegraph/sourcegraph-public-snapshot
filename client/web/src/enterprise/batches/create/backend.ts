import { gql } from '@sourcegraph/http-client'

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
            id
        }
        description

        currentSpec {
            id
            originalInput
            createdAt
        }

        batchSpecs(first: 1) {
            nodes {
                id
                originalInput
                createdAt
            }
        }

        state
    }
`

export const EXECUTE_BATCH_SPEC = gql`
    mutation ExecuteBatchSpec($batchSpec: ID!) {
        executeBatchSpec(batchSpec: $batchSpec) {
            id
            namespace {
                url
            }
        }
    }
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
    mutation CreateBatchSpecFromRaw($spec: String!, $noCache: Boolean!, $namespace: ID!) {
        createBatchSpecFromRaw(batchSpec: $spec, noCache: $noCache, namespace: $namespace) {
            id
            createdAt
        }
    }
`

// This mutation is used to update the batch spec when the existing batch spec is
// unapplied.
export const REPLACE_BATCH_SPEC_INPUT = gql`
    mutation ReplaceBatchSpecInput($previousSpec: ID!, $spec: String!, $noCache: Boolean!) {
        replaceBatchSpecInput(previousSpec: $previousSpec, batchSpec: $spec, noCache: $noCache) {
            id
            createdAt
        }
    }
`

export const WORKSPACE_RESOLUTION_STATUS = gql`
    query WorkspaceResolutionStatus($batchSpec: ID!) {
        node(id: $batchSpec) {
            __typename
            ... on BatchSpec {
                workspaceResolution {
                    state
                    failureMessage
                }
            }
        }
    }
`

export const WORKSPACES = gql`
    query BatchSpecWorkspacesPreview($batchSpec: ID!, $first: Int, $after: String) {
        node(id: $batchSpec) {
            __typename
            ... on BatchSpec {
                workspaceResolution {
                    __typename
                    workspaces(first: $first, after: $after) {
                        __typename
                        totalCount
                        pageInfo {
                            hasNextPage
                            endCursor
                        }
                        nodes {
                            ...PreviewBatchSpecWorkspaceFields
                        }
                    }
                }
            }
        }
    }

    fragment PreviewBatchSpecWorkspaceFields on BatchSpecWorkspace {
        __typename
        repository {
            __typename
            id
            name
            url
        }
        ignored
        unsupported
        branch {
            __typename
            id
            abbrevName
            displayName
            target {
                __typename
                oid
            }
            url
        }
        path
        searchResultPaths
        cachedResultFound
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
