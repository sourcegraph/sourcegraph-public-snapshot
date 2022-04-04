import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { fileDiffFields } from '../../../backend/diff'
import { requestGraphQL } from '../../../backend/graphql'
import { useConnection, UseConnectionResult } from '../../../components/FilteredConnection/hooks/useConnection'
import {
    BatchSpecExecutionByIDResult,
    BatchSpecExecutionByIDVariables,
    BatchSpecExecutionFields,
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
    RetryWorkspaceExecutionResult,
    RetryWorkspaceExecutionVariables,
    RetryBatchSpecExecutionResult,
    RetryBatchSpecExecutionVariables,
    BatchSpecWorkspaceState,
    VisibleBatchSpecWorkspaceFields,
    HiddenBatchSpecWorkspaceFields,
    VisibleBatchSpecWorkspaceListFields,
    HiddenBatchSpecWorkspaceListFields,
} from '../../../graphql-operations'

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

const batchSpecExecutionFieldsFragment = gql`
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
                    errored
                    completed
                    processing
                    queued
                    ignored
                }
            }
        }
    }
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

export const fetchBatchSpecExecution = (id: Scalars['ID']): Observable<BatchSpecExecutionFields | null> =>
    requestGraphQL<BatchSpecExecutionByIDResult, BatchSpecExecutionByIDVariables>(FETCH_BATCH_SPEC_EXECUTION, {
        id,
    }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'BatchSpec') {
                throw new Error(`Node is a ${node.__typename}, not a BatchSpec`)
            }
            return node
        })
    )

export const fetchBatchSpecWorkspace = (
    id: Scalars['ID']
): Observable<HiddenBatchSpecWorkspaceFields | VisibleBatchSpecWorkspaceFields | null> =>
    requestGraphQL<BatchSpecWorkspaceByIDResult, BatchSpecWorkspaceByIDVariables>(
        gql`
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
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'HiddenBatchSpecWorkspace' && node.__typename !== 'VisibleBatchSpecWorkspace') {
                throw new Error(`Node is a ${node.__typename}, not a BatchSpecWorkspace`)
            }
            return node
        })
    )

export async function cancelBatchSpecExecution(id: Scalars['ID']): Promise<BatchSpecExecutionFields> {
    const result = await requestGraphQL<CancelBatchSpecExecutionResult, CancelBatchSpecExecutionVariables>(
        gql`
            mutation CancelBatchSpecExecution($id: ID!) {
                cancelBatchSpecExecution(batchSpec: $id) {
                    ...BatchSpecExecutionFields
                }
            }

            ${batchSpecExecutionFieldsFragment}
        `,
        { id }
    ).toPromise()
    return dataOrThrowErrors(result).cancelBatchSpecExecution
}

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
                throw new Error(`The given ID is a ${node.__typename}, not a BatchSpecWorkspace`)
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

const BATCH_SPEC_WORKSPACES = gql`
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
            // For some reason caching caused flickering here. Will need to investigate
            // further why.
            fetchPolicy: 'no-cache',
            pollInterval: 1000,
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

export async function retryWorkspaceExecution(id: Scalars['ID']): Promise<void> {
    const result = await requestGraphQL<RetryWorkspaceExecutionResult, RetryWorkspaceExecutionVariables>(
        gql`
            mutation RetryWorkspaceExecution($id: ID!) {
                retryBatchSpecWorkspaceExecution(batchSpecWorkspaces: [$id]) {
                    alwaysNil
                }
            }
        `,
        { id }
    ).toPromise()
    dataOrThrowErrors(result)
}

export async function retryBatchSpecExecution(id: Scalars['ID']): Promise<BatchSpecExecutionFields> {
    return requestGraphQL<RetryBatchSpecExecutionResult, RetryBatchSpecExecutionVariables>(
        gql`
            mutation RetryBatchSpecExecution($id: ID!) {
                retryBatchSpecExecution(batchSpec: $id) {
                    ...BatchSpecExecutionFields
                }
            }

            ${batchSpecExecutionFieldsFragment}
        `,
        { id }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(({ retryBatchSpecExecution }) => retryBatchSpecExecution)
        )
        .toPromise()
}
