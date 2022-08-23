import { MutationTuple, useApolloClient, gql as apolloGQL } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { asError, ErrorLike } from '@sourcegraph/common'
import { dataOrThrowErrors, gql, useMutation, useQuery } from '@sourcegraph/http-client'
import { useStopwatch } from '@sourcegraph/wildcard'

import { fileDiffFields } from '../../../../backend/diff'
import { requestGraphQL } from '../../../../backend/graphql'
import { useConnection, UseConnectionResult } from '../../../../components/FilteredConnection/hooks/useConnection'
import {
    BatchSpecWorkspaceAndStatusFields,
    BatchSpecWorkspaceByIDResult,
    BatchSpecWorkspaceByIDVariables,
    BatchSpecWorkspaceStatusesConnectionFields,
    BatchSpecWorkspacesResult,
    BatchSpecWorkspaceState,
    BatchSpecWorkspaceStatusesResult,
    BatchSpecWorkspaceStatusesVariables,
    BatchSpecWorkspaceStepFileDiffsResult,
    BatchSpecWorkspaceStepFileDiffsVariables,
    BatchSpecWorkspacesVariables,
    CancelBatchSpecExecutionResult,
    CancelBatchSpecExecutionVariables,
    HiddenBatchSpecWorkspaceFields,
    HiddenBatchSpecWorkspaceListFields,
    RetryWorkspaceExecutionResult,
    RetryWorkspaceExecutionVariables,
    Scalars,
    VisibleBatchSpecWorkspaceFields,
    VisibleBatchSpecWorkspaceListFields,
    WorkspaceStepFileDiffConnectionFields,
} from '../../../../graphql-operations'

const WORKSPACE_POLLING_INTERVAL = 2500

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
            __typename
            id
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
            pollInterval: WORKSPACE_POLLING_INTERVAL,
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

/**
 * BATCH_SPEC_WORKSPACE_STATUSES is the GraphQL document to query a batch spec's
 * workspaces connection for dynamic fields that indicate the workspace's execution
 * status, namely: `placeInQueue`, `state`, and `diffStat`. These fields are available on
 * both visible and hidden workspaces.
 *
 * Due to the `defaultMaxFirstParam` enforced by the backend server, we can only request
 * up to 10000 workspaces at a time.
 */
export const BATCH_SPEC_WORKSPACE_STATUSES = gql`
    query BatchSpecWorkspaceStatuses($node: ID!, $after: String) {
        node(id: $node) {
            __typename
            ... on BatchSpec {
                id
                workspaceResolution {
                    __typename
                    id
                    workspaces(first: 10000, after: $after) {
                        ... on BatchSpecWorkspaceConnection {
                            ...BatchSpecWorkspaceStatusesConnectionFields
                        }
                    }
                }
            }
        }
    }

    fragment BatchSpecWorkspaceStatusesConnectionFields on BatchSpecWorkspaceConnection {
        __typename
        pageInfo {
            hasNextPage
            endCursor
        }
        nodes {
            __typename
            ... on BatchSpecWorkspace {
                ...BatchSpecWorkspaceStatusFields
            }
        }
    }

    fragment BatchSpecWorkspaceStatusFields on BatchSpecWorkspace {
        __typename
        id
        state
        diffStat {
            added
            changed
            deleted
        }
        placeInQueue
    }
`

/**
 * workspacesOrError is a helper function that extracts the `BatchSpecWorkspaceConnection`
 * from the full query response data, or throws an error if it's not accessible in the
 * response.
 */
const workspacesOrError = (
    batchSpecID: Scalars['ID'],
    data: BatchSpecWorkspaceStatusesResult
): BatchSpecWorkspaceStatusesConnectionFields | null => {
    if (!data.node) {
        throw new Error(`Batch spec with ID ${batchSpecID} does not exist`)
    }
    if (data.node.__typename !== 'BatchSpec') {
        throw new Error(`The given ID is a ${data.node.__typename as string}, not a BatchSpec`)
    }
    return data.node.workspaceResolution?.workspaces || null
}

/**
 * usePollWorkspaceStatuses is a hook that will poll batched requests for a batch spec's
 * workspaces connection for dynamic fields that indicate every workspace's execution
 * status.
 *
 * It is meant to be used in conjunction with `useWorkspacesListConnection` to populate
 * all of the workspace information surfaced in the UI for the execution workspaces list,
 * but it is separate in order to prevent write conflicts between `fetchMore` requests
 * (the "show more" button) and polling requests. The data from both queries is merged and
 * eventually read from the cache.
 */
export const usePollWorkspaceStatuses = (batchSpecID: Scalars['ID']): void => {
    const { start: resetStopwatch, time } = useStopwatch()
    // We manually poll with `refetch` in case there are more workspaces than fit on the
    // first page.
    const { refetch } = useQuery<BatchSpecWorkspaceStatusesResult, BatchSpecWorkspaceStatusesVariables>(
        BATCH_SPEC_WORKSPACE_STATUSES,
        {
            variables: { node: batchSpecID, after: null },
            fetchPolicy: 'cache-and-network',
            /* eslint-disable @typescript-eslint/no-floating-promises */
            onCompleted: data => {
                const workspaces = workspacesOrError(batchSpecID, data)
                if (!workspaces) {
                    return
                }
                // If there are more workspaces to fetch, do that immediately.
                if (workspaces.pageInfo.hasNextPage) {
                    refetch({ node: batchSpecID, after: workspaces.pageInfo.endCursor })
                    return
                }
                // Once we've finished fetching all of the workspace statuses, we'll wait
                // the extent of the polling interval and then start back at the beginning
                // again. Ideally, we would stop polling once all workspaces were in a
                // final state. However, we don't have an easy way to restart polling if
                // the user then retried any workspace, so for this reason we won't stop.
                if (time.milliseconds > WORKSPACE_POLLING_INTERVAL) {
                    // If it took longer than the polling interval to fetch all of the
                    // pages of workspaces, we can restart immediately.
                    resetStopwatch()
                    refetch({ node: batchSpecID, after: null })
                } else {
                    // Otherwise, we'll wait for the remaining part of the interval to
                    // elapse first.
                    setTimeout(() => {
                        resetStopwatch()
                        refetch({ node: batchSpecID, after: null })
                    }, WORKSPACE_POLLING_INTERVAL - time.milliseconds)
                }
            },
            /* eslint-enable @typescript-eslint/no-floating-promises */
        }
    )
}

const commonBatchSpecWorkspaceFields = gql`
    fragment BatchSpecWorkspaceListFields on BatchSpecWorkspace {
        __typename
        id
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

/**
 * BATCH_SPEC_WORKSPACES is the GraphQL document to query a batch spec's workspaces
 * connection for fields of the workspaces that are not expected to change, such as the
 * path/branch and repository it is part of or whether or not it's
 * unsupported/ignored/cached.
 */
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
                    __typename
                    id
                    workspaces(first: $first, after: $after, search: $search, state: $state) {
                        ...BatchSpecWorkspaceConnectionFields
                    }
                }
            }
        }
    }

    fragment BatchSpecWorkspaceConnectionFields on BatchSpecWorkspaceConnection {
        __typename
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

    ${commonBatchSpecWorkspaceFields}
`

/**
 * usePollWorkspaceStatuses is a connection query hook for a batch spec's workspaces
 * connection, specifically for fields of the workspaces that are not expected to change.
 * It can optionally filter the nodes in the connection by search string or execution
 * state.
 *
 * It is meant to be used in conjunction with `usePollWorkspaceStatuses` to populate all
 * of the workspace information surfaced in the UI for the execution workspaces list.
 */
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
            first: 50,
            search,
            state,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
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

/**
 * useWorkspaceFromCache is a hook that fetches a given batch spec workspace from the
 * Apollo Client cache, if it is there. It will *not* use the network at all and relies on
 * data having already been fetched and written from `useWorkspacesListConnection` and
 * `usePollWorkspaceStatuses`. It is intended to be used to fetch the merged results from
 * both of said queries.
 */
export const useWorkspaceFromCache = (
    workspaceID: Scalars['ID'],
    type: 'VisibleBatchSpecWorkspace' | 'HiddenBatchSpecWorkspace'
): BatchSpecWorkspaceAndStatusFields | null | undefined => {
    const apolloClient = useApolloClient()
    return apolloClient.readFragment<BatchSpecWorkspaceAndStatusFields | null | undefined>({
        id: `${type}:${workspaceID}`,
        fragmentName: 'BatchSpecWorkspaceAndStatusFields',
        fragment: apolloGQL`
            fragment BatchSpecWorkspaceAndStatusFields on BatchSpecWorkspace {
                __typename
                id
                state
                diffStat {
                    added
                    changed
                    deleted
                }
                placeInQueue
                __typename
                ... on HiddenBatchSpecWorkspace {
                    ...HiddenBatchSpecWorkspaceListFields
                }
                ... on VisibleBatchSpecWorkspace {
                    ...VisibleBatchSpecWorkspaceListFields
                }
            }

            ${commonBatchSpecWorkspaceFields}
        `,
    })
}

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
