import { MutationTuple } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { asError, ErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors, gql, useMutation, useQuery } from '@sourcegraph/http-client'

import { fileDiffFields } from '../../../../backend/diff'
import { requestGraphQL } from '../../../../backend/graphql'
import { useConnection, UseConnectionResult } from '../../../../components/FilteredConnection/hooks/useConnection'
import {
    BatchSpecWorkspaceByIDResult,
    BatchSpecWorkspaceByIDVariables,
    BatchSpecWorkspacesResult,
    BatchSpecWorkspaceStepFileDiffsResult,
    BatchSpecWorkspaceStepFileDiffsVariables,
    BatchSpecWorkspacesVariables,
    CancelBatchSpecExecutionResult,
    CancelBatchSpecExecutionVariables,
    Scalars,
    WorkspaceStepFileDiffConnectionFields,
    BatchSpecWorkspaceState,
    VisibleBatchSpecWorkspaceFields,
    HiddenBatchSpecWorkspaceFields,
    VisibleBatchSpecWorkspaceListFields,
    HiddenBatchSpecWorkspaceListFields,
    RetryWorkspaceExecutionResult,
    RetryWorkspaceExecutionVariables,
} from '../../../../graphql-operations'

const batchSpecWorkspaceFieldsFragment = gql`
    fragment BatchSpecWorkspaceFields on BatchSpecWorkspace {
        __typename
        id
        queuedAt
        startedAt
        finishedAt
        state
        diffStat {
            added
            changed
            deleted
        }
        placeInQueue
        onlyFetchWorkspace
        ignored
        unsupported
        cachedResultFound
    }

    fragment VisibleBatchSpecWorkspaceFields on VisibleBatchSpecWorkspace {
        ...BatchSpecWorkspaceFields
        steps {
            ...BatchSpecWorkspaceStepFields
        }
        searchResultPaths
        failureMessage
        changesetSpecs {
            __typename
            ...BatchSpecWorkspaceChangesetSpecFields
        }
        repository {
            name
            url
        }
        branch {
            displayName
        }
        path
        stages {
            setup {
                ...BatchSpecWorkspaceExecutionLogEntryFields
            }
            srcExec {
                ...BatchSpecWorkspaceExecutionLogEntryFields
            }
            teardown {
                ...BatchSpecWorkspaceExecutionLogEntryFields
            }
        }
        executor {
            __typename
            id
            queueName
            hostname
            active
            os
            architecture
            dockerVersion
            executorVersion
            gitVersion
            igniteVersion
            srcCliVersion
            firstSeenAt
            lastSeenAt
        }
    }

    fragment HiddenBatchSpecWorkspaceFields on HiddenBatchSpecWorkspace {
        ...BatchSpecWorkspaceFields
    }

    fragment BatchSpecWorkspaceStepFields on BatchSpecWorkspaceStep {
        number
        run
        diffStat {
            added
            changed
            deleted
        }
        container
        ifCondition
        cachedResultFound
        skipped
        outputLines
        startedAt
        finishedAt
        exitCode
        environment {
            name
            value
        }
        outputVariables {
            name
            value
        }
    }

    fragment BatchSpecWorkspaceExecutionLogEntryFields on ExecutionLogEntry {
        key
        command
        startTime
        exitCode
        out
        durationMilliseconds
    }

    fragment BatchSpecWorkspaceChangesetSpecFields on ChangesetSpec {
        id
        type
        __typename
        ... on VisibleChangesetSpec {
            description {
                __typename
                ... on GitBranchChangesetDescription {
                    title
                    body
                    baseRepository {
                        name
                        url
                    }
                    published
                    baseRef
                    headRef
                    diffStat {
                        added
                        changed
                        deleted
                    }
                }
            }
        }
    }
`

const batchSpecWorkspaceStatsFragment = gql`
    fragment BatchSpecWorkspaceStats on BatchSpecWorkspacesStats {
        errored
        completed
        processing
        queued
        ignored
    }
`

export const batchSpecExecutionFieldsFragment = gql`
    fragment BatchSpecExecutionFields on BatchSpec {
        id
        originalInput
        state
        description {
            name
        }
        createdAt
        startedAt
        finishedAt
        failureMessage
        applyURL
        creator {
            id
            url
            displayName
            username
        }
        namespace {
            id
            url
            namespaceName
        }
        appliesToBatchChange {
            url
        }
        viewerCanRetry
        workspaceResolution {
            workspaces {
                stats {
                    ...BatchSpecWorkspaceStats
                }
            }
        }
    }

    ${batchSpecWorkspaceStatsFragment}
`

export const FETCH_BATCH_SPEC_EXECUTION = gql`
    query BatchSpecExecutionByID($id: ID!) {
        node(id: $id) {
            __typename
            ...BatchSpecExecutionFields
        }
    }

    ${batchSpecExecutionFieldsFragment}
`

export const BATCH_SPEC_WORKSPACE_BY_ID = gql`
    query BatchSpecWorkspaceByID($id: ID!) {
        node(id: $id) {
            __typename
            ... on HiddenBatchSpecWorkspace {
                ...HiddenBatchSpecWorkspaceFields
            }
            ... on VisibleBatchSpecWorkspace {
                ...VisibleBatchSpecWorkspaceFields
            }
        }
    }
    ${batchSpecWorkspaceFieldsFragment}
`

interface BatchSpecWorkspaceHookResult {
    data?: VisibleBatchSpecWorkspaceFields | HiddenBatchSpecWorkspaceFields | null
    error?: ErrorLike
    loading: boolean
}

export const useBatchSpecWorkspace = (id: Scalars['ID']): BatchSpecWorkspaceHookResult => {
    const { loading, data, error } = useQuery<BatchSpecWorkspaceByIDResult, BatchSpecWorkspaceByIDVariables>(
        BATCH_SPEC_WORKSPACE_BY_ID,
        {
            variables: { id },
            // Cache this data but always re-request it in the background to pick up newer changes.
            fetchPolicy: 'cache-and-network',
            // We continuously poll for changes to the workspace. This isn't the most effective
            // use of network bandwidth since many of these fields aren't changing and most of
            // the time there will be no changes at all, but it's also the easiest way to
            // keep this in sync for now at the cost of a bit of excess network resources.
            pollInterval: 2500,
        }
    )

    const result: BatchSpecWorkspaceHookResult = {
        loading,
        error: error ? asError(error) : undefined,
    }

    if (data?.node) {
        if (
            data.node.__typename !== 'HiddenBatchSpecWorkspace' &&
            data.node.__typename !== 'VisibleBatchSpecWorkspace'
        ) {
            throw new Error(`Node is a ${data.node.__typename}, not a BatchSpecWorkspace`)
        }
        result.data = data.node
    }

    return result
}

export const CANCEL_BATCH_SPEC_EXECUTION = gql`
    mutation CancelBatchSpecExecution($id: ID!) {
        cancelBatchSpecExecution(batchSpec: $id) {
            ...BatchSpecExecutionFields
        }
    }

    ${batchSpecExecutionFieldsFragment}
`

export const useCancelBatchSpecExecution = (
    batchSpecID: Scalars['ID']
): MutationTuple<CancelBatchSpecExecutionResult, CancelBatchSpecExecutionVariables> =>
    useMutation(CANCEL_BATCH_SPEC_EXECUTION, { variables: { id: batchSpecID } })

const batchSpecWorkspaceStepFileDiffsFields = gql`
    fragment BatchSpecWorkspaceStepFileDiffsFields on VisibleBatchSpecWorkspace {
        step(index: $step) {
            diff {
                fileDiffs(first: $first, after: $after) {
                    ...WorkspaceStepFileDiffConnectionFields
                }
            }
        }
    }

    fragment WorkspaceStepFileDiffConnectionFields on FileDiffConnection {
        nodes {
            ...FileDiffFields
        }
        totalCount
        pageInfo {
            hasNextPage
            endCursor
        }
    }

    ${fileDiffFields}
`

// TODO: `FileDiffConnection` is implemented with observables and expects this query to be
// provided as one, so we can't migrate this to Apollo Client yet.
export const queryBatchSpecWorkspaceStepFileDiffs = ({
    node: nodeID,
    step,
    first,
    after,
}: BatchSpecWorkspaceStepFileDiffsVariables): Observable<WorkspaceStepFileDiffConnectionFields> =>
    requestGraphQL<BatchSpecWorkspaceStepFileDiffsResult, BatchSpecWorkspaceStepFileDiffsVariables>(
        gql`
            query BatchSpecWorkspaceStepFileDiffs($node: ID!, $step: Int!, $first: Int, $after: String) {
                node(id: $node) {
                    __typename
                    ... on VisibleBatchSpecWorkspace {
                        ...BatchSpecWorkspaceStepFileDiffsFields
                    }
                }
            }

            ${batchSpecWorkspaceStepFileDiffsFields}
        `,
        { node: nodeID, step, first, after }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`BatchSpecWorkspace with ID ${nodeID} does not exist`)
            }
            if (node.__typename === 'HiddenBatchSpecWorkspace') {
                throw new Error('No access to this workspace')
            }
            if (node.__typename !== 'VisibleBatchSpecWorkspace') {
                throw new Error(`The given ID is a ${node.__typename}, not a VisibleBatchSpecWorkspace`)
            }
            if (!node.step) {
                throw new Error('The given Step is not available')
            }
            if (!node.step.diff) {
                throw new Error('The diff is not available')
            }
            return node.step.diff.fileDiffs
        })
    )

export const BATCH_SPEC_WORKSPACES = gql`
    query BatchSpecWorkspaces(
        $node: ID!
        $first: Int
        $after: String
        $search: String
        $state: BatchSpecWorkspaceState
    ) {
        node(id: $node) {
            __typename
            ... on BatchSpec {
                id
                workspaceResolution {
                    workspaces(first: $first, after: $after, search: $search, state: $state) {
                        ...BatchSpecWorkspacesConnectionFields
                    }
                }
            }
        }
    }

    fragment BatchSpecWorkspacesConnectionFields on BatchSpecWorkspaceConnection {
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
        nodes {
            __typename
            ... on HiddenBatchSpecWorkspace {
                ...HiddenBatchSpecWorkspaceListFields
            }
            ... on VisibleBatchSpecWorkspace {
                ...VisibleBatchSpecWorkspaceListFields
            }
        }
    }

    fragment BatchSpecWorkspaceListFields on BatchSpecWorkspace {
        __typename
        id
        state
        diffStat {
            added
            changed
            deleted
        }
        placeInQueue
        ignored
        unsupported
        cachedResultFound
    }

    fragment VisibleBatchSpecWorkspaceListFields on VisibleBatchSpecWorkspace {
        __typename
        ...BatchSpecWorkspaceListFields
        repository {
            name
            url
        }
        branch {
            displayName
        }
        path
    }

    fragment HiddenBatchSpecWorkspaceListFields on HiddenBatchSpecWorkspace {
        __typename
        ...BatchSpecWorkspaceListFields
    }
`

export const useWorkspacesListConnection = (
    batchSpecID: Scalars['ID'],
    search: string | null,
    state: BatchSpecWorkspaceState | null
): UseConnectionResult<HiddenBatchSpecWorkspaceListFields | VisibleBatchSpecWorkspaceListFields> =>
    useConnection<
        BatchSpecWorkspacesResult,
        BatchSpecWorkspacesVariables,
        HiddenBatchSpecWorkspaceListFields | VisibleBatchSpecWorkspaceListFields
    >({
        query: BATCH_SPEC_WORKSPACES,
        variables: {
            node: batchSpecID,
            after: null,
            first: 20,
            search,
            state,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
            pollInterval: 2500,
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)

            if (!data.node) {
                throw new Error(`Batch spec with ID ${batchSpecID} does not exist`)
            }
            if (data.node.__typename !== 'BatchSpec') {
                throw new Error(`The given ID is a ${data.node.__typename as string}, not a BatchSpec`)
            }
            if (!data.node.workspaceResolution) {
                throw new Error('No workspace resolution in batch spec')
            }
            return data.node.workspaceResolution.workspaces
        },
    })

const RETRY_WORKSPACE_EXECUTION = gql`
    mutation RetryWorkspaceExecution($id: ID!) {
        retryBatchSpecWorkspaceExecution(batchSpecWorkspaces: [$id]) {
            alwaysNil
        }
    }
`
export const useRetryWorkspaceExecution = (
    workspaceID: Scalars['ID']
): MutationTuple<RetryWorkspaceExecutionResult, RetryWorkspaceExecutionVariables> =>
    useMutation(RETRY_WORKSPACE_EXECUTION, { variables: { id: workspaceID } })

export const RETRY_BATCH_SPEC_EXECUTION = gql`
    mutation RetryBatchSpecExecution($id: ID!) {
        retryBatchSpecExecution(batchSpec: $id) {
            ...BatchSpecExecutionFields
        }
    }

    ${batchSpecExecutionFieldsFragment}
`
