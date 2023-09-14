import type { MutationTuple } from '@apollo/client'
import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { asError, type ErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors, gql, useMutation, useQuery } from '@sourcegraph/http-client'

import { fileDiffFields } from '../../../../backend/diff'
import { requestGraphQL } from '../../../../backend/graphql'
import type {
    BatchSpecWorkspaceByIDResult,
    BatchSpecWorkspaceByIDVariables,
    BatchSpecWorkspacesConnectionFields,
    BatchSpecWorkspacesResult,
    BatchSpecWorkspaceStepFileDiffsResult,
    BatchSpecWorkspaceStepFileDiffsVariables,
    BatchSpecWorkspacesVariables,
    CancelBatchSpecExecutionResult,
    CancelBatchSpecExecutionVariables,
    Scalars,
    WorkspaceStepFileDiffConnectionFields,
    VisibleBatchSpecWorkspaceFields,
    HiddenBatchSpecWorkspaceFields,
    RetryWorkspaceExecutionResult,
    RetryWorkspaceExecutionVariables,
} from '../../../../graphql-operations'

export const batchSpecWorkspaceStepOutputLinesFieldsFragment = gql`
    fragment BatchSpecWorkspaceStepOutputLines on BatchSpecWorkspaceStepOutputLineConnection {
        __typename
        nodes
        totalCount
        pageInfo {
            hasNextPage
            endCursor
        }
    }
`

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
            deleted
        }
        placeInQueue
        placeInGlobalQueue
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
            queueNames
            hostname
            active
            os
            compatibility
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
            deleted
        }
        container
        ifCondition
        cachedResultFound
        skipped
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
        source
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

export const BATCH_SPEC_WORKSPACE_STEP = gql`
    query BatchSpecWorkspaceStep($workspaceID: ID!, $stepIndex: Int!, $first: Int!, $after: String) {
        node(id: $workspaceID) {
            __typename
            ... on VisibleBatchSpecWorkspace {
                step(index: $stepIndex) {
                    outputLines(first: $first, after: $after) {
                        ...BatchSpecWorkspaceStepOutputLines
                    }
                }
            }
        }
    }

    ${batchSpecWorkspaceStepOutputLinesFieldsFragment}
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
        $batchSpecID: ID!
        $first: Int
        $after: String
        $search: String
        $state: BatchSpecWorkspaceState
    ) {
        node(id: $batchSpecID) {
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

// NOTE: The workspaces list connection query was implemented with Apollo but has been
// migrated back to `requestGraphQL` and observables due to polling + pagination not
// playing well together with the cache. See
// https://github.com/sourcegraph/sourcegraph/pull/40717 for more context. 😢😢
export const queryWorkspacesList = ({
    batchSpecID,
    first,
    after,
    search,
    state,
}: BatchSpecWorkspacesVariables): Observable<BatchSpecWorkspacesConnectionFields> =>
    requestGraphQL<BatchSpecWorkspacesResult, BatchSpecWorkspacesVariables>(BATCH_SPEC_WORKSPACES, {
        batchSpecID,
        first,
        after,
        search,
        state,
    }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error(`Batch spec with ID ${batchSpecID} does not exist`)
            }
            if (node.__typename !== 'BatchSpec') {
                throw new Error(`The given ID is a ${node.__typename}, not a BatchSpec`)
            }
            if (!node.workspaceResolution) {
                throw new Error('No workspace resolution in batch spec')
            }
            return node.workspaceResolution.workspaces
        })
    )

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
