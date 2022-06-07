import { subDays } from 'date-fns'

import {
    BatchChangeFields,
    BulkOperationState,
    BulkOperationType,
    BatchChangeBulkOperationsResult,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
    BatchChangeChangesetsResult,
    ChangesetCheckState,
    BatchSpecState,
    BatchChangeState,
} from '../../../graphql-operations'

const now = new Date()

export const MOCK_BATCH_CHANGE: BatchChangeFields = {
    __typename: 'BatchChange',
    state: BatchChangeState.OPEN,
    changesetsStats: {
        __typename: 'ChangesetsStats',
        closed: 1,
        deleted: 1,
        draft: 1,
        merged: 2,
        open: 2,
        archived: 5,
        total: 18,
        unpublished: 4,
    },
    createdAt: subDays(now, 5).toISOString(),
    creator: {
        url: '/users/alice',
        username: 'alice',
    },
    id: 'specid',
    url: '/users/alice/batch-changes/awesome-batch-change',
    namespace: {
        id: '1234',
        namespaceName: 'alice',
        url: '/users/alice',
    },
    viewerCanAdminister: true,
    closedAt: null,
    description: '## What this batch change does\n\nTruly awesome things for example.',
    name: 'awesome-batch-changes',
    updatedAt: subDays(now, 5).toISOString(),
    lastAppliedAt: subDays(now, 5).toISOString(),
    lastApplier: {
        url: '/users/bob',
        username: 'bob',
    },
    currentSpec: {
        id: 'specID1',
        originalInput: 'name: awesome-batch-changes\ndescription: somestring',
        supersedingBatchSpec: null,
        codeHostsWithoutWebhooks: {
            nodes: [],
            pageInfo: { hasNextPage: false },
            totalCount: 0,
        },
    },
    batchSpecs: {
        nodes: [{ state: BatchSpecState.COMPLETED }],
        pageInfo: { hasNextPage: false },
    },
    bulkOperations: {
        __typename: 'BulkOperationConnection',
        totalCount: 3,
    },
    activeBulkOperations: {
        __typename: 'BulkOperationConnection',
        totalCount: 1,
        nodes: [
            {
                __typename: 'BulkOperation',
                id: 'testid-123',
                state: BulkOperationState.PROCESSING,
            },
        ],
    },
    diffStat: { added: 1000, changed: 2000, deleted: 1000, __typename: 'DiffStat' },
}

export const MOCK_BULK_OPERATIONS: BatchChangeBulkOperationsResult = {
    node: {
        __typename: 'BatchChange',
        bulkOperations: {
            __typename: 'BulkOperationConnection',
            totalCount: 3,
            pageInfo: {
                endCursor: null,
                hasNextPage: false,
            },
            nodes: [
                {
                    __typename: 'BulkOperation',
                    id: 'id1',
                    type: BulkOperationType.COMMENT,
                    state: BulkOperationState.PROCESSING,
                    errors: [],
                    progress: 0.25,
                    createdAt: subDays(now, 5).toISOString(),
                    finishedAt: null,
                    changesetCount: 100,
                    initiator: {
                        url: '/users/alice',
                        username: 'alice',
                    },
                },
                {
                    __typename: 'BulkOperation',
                    id: 'id2',
                    type: BulkOperationType.COMMENT,
                    state: BulkOperationState.COMPLETED,
                    errors: [],
                    progress: 1,
                    createdAt: subDays(now, 5).toISOString(),
                    finishedAt: subDays(now, 4).toISOString(),
                    changesetCount: 100,
                    initiator: {
                        url: '/users/alice',
                        username: 'alice',
                    },
                },
                {
                    __typename: 'BulkOperation',
                    id: 'id3',
                    type: BulkOperationType.DETACH,
                    state: BulkOperationState.COMPLETED,
                    errors: [],
                    progress: 1,
                    createdAt: subDays(now, 5).toISOString(),
                    finishedAt: subDays(now, 4).toISOString(),
                    changesetCount: 25,
                    initiator: {
                        url: '/users/alice',
                        username: 'alice',
                    },
                },
                {
                    __typename: 'BulkOperation',
                    id: 'id4',
                    type: BulkOperationType.COMMENT,
                    state: BulkOperationState.FAILED,
                    errors: [
                        {
                            changeset: {
                                __typename: 'ExternalChangeset',
                                externalURL: {
                                    url: 'https://test.test/my/pr',
                                },
                                repository: {
                                    name: 'sourcegraph/sourcegraph',
                                    url: '/github.com/sourcegraph/sourcegraph',
                                },
                                title: 'Changeset title on code host',
                            },
                            error: 'Failed to create comment, cannot comment on a PR that is awesome.',
                        },
                    ],
                    progress: 1,
                    createdAt: subDays(now, 5).toISOString(),
                    finishedAt: subDays(now, 4).toISOString(),
                    changesetCount: 100,
                    initiator: {
                        url: '/users/alice',
                        username: 'alice',
                    },
                },
            ],
        },
    },
}

export const BATCH_CHANGE_CHANGESETS_RESULT: BatchChangeChangesetsResult['node'] = {
    ...MOCK_BATCH_CHANGE,
    changesets: {
        __typename: 'ChangesetConnection',
        totalCount: 7,
        nodes: [
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.UNPUBLISHED,
                id: 'someh1',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.PROCESSING,
                id: 'someh2',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.RETRYING,
                id: 'someh3',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.FAILED,
                id: 'someh4',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.OPEN,
                id: 'someh5',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'ExternalChangeset',
                body: 'body',
                checkState: ChangesetCheckState.PASSED,
                diffStat: {
                    __typename: 'DiffStat',
                    added: 10,
                    changed: 9,
                    deleted: 1,
                },
                externalID: '123',
                externalURL: {
                    url: 'http://test.test/123',
                },
                forkNamespace: null,
                labels: [
                    {
                        __typename: 'ChangesetLabel',
                        color: '93ba13',
                        description: 'Very awesome description',
                        text: 'Some label',
                    },
                ],
                repository: {
                    id: 'repoid',
                    name: 'github.com/sourcegraph/awesome',
                    url: 'http://test.test/awesome',
                },
                reviewState: ChangesetReviewState.COMMENTED,
                title: 'Add prettier to all projects',
                createdAt: subDays(now, 5).toISOString(),
                updatedAt: subDays(now, 5).toISOString(),
                state: ChangesetState.OPEN,
                nextSyncAt: null,
                id: 'somev1',
                error: null,
                syncerError: null,
                currentSpec: {
                    id: 'spec-rand-id-1',
                    type: ChangesetSpecType.BRANCH,
                    description: {
                        __typename: 'GitBranchChangesetDescription',
                        baseRef: 'my-branch',
                        headRef: 'my-branch',
                    },
                    forkTarget: null,
                },
            },
            {
                __typename: 'ExternalChangeset',
                body: 'body',
                checkState: null,
                diffStat: {
                    __typename: 'DiffStat',
                    added: 10,
                    changed: 9,
                    deleted: 1,
                },
                externalID: null,
                externalURL: null,
                forkNamespace: null,
                labels: [],
                repository: {
                    id: 'repoid',
                    name: 'github.com/sourcegraph/awesome',
                    url: 'http://test.test/awesome',
                },
                reviewState: null,
                title: 'Add prettier to all projects',
                createdAt: subDays(now, 5).toISOString(),
                updatedAt: subDays(now, 5).toISOString(),
                state: ChangesetState.RETRYING,
                nextSyncAt: null,
                id: 'somev2',
                error: 'Cannot create PR, insufficient token scope.',
                syncerError: null,
                currentSpec: {
                    id: 'spec-rand-id-2',
                    type: ChangesetSpecType.BRANCH,
                    description: {
                        __typename: 'GitBranchChangesetDescription',
                        baseRef: 'my-branch',
                        headRef: 'my-branch',
                    },
                    forkTarget: null,
                },
            },
        ],
        pageInfo: { endCursor: null, hasNextPage: false },
    },
}

export const EMPTY_BATCH_CHANGE_CHANGESETS_RESULT: BatchChangeChangesetsResult['node'] = {
    ...MOCK_BATCH_CHANGE,
    changesets: {
        __typename: 'ChangesetConnection',
        totalCount: 0,
        nodes: [],
        pageInfo: { endCursor: null, hasNextPage: false },
    },
}
