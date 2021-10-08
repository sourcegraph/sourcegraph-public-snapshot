import { subDays } from 'date-fns'

import { BatchChangeFields, BulkOperationState } from '../../../graphql-operations'

const now = new Date()

export const MOCK_BATCH_CHANGE: BatchChangeFields = {
    __typename: 'BatchChange',
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
    initialApplier: {
        url: '/users/alice',
        username: 'alice',
    },
    id: 'specid',
    url: '/users/alice/batch-changes/awesome-batch-change',
    namespace: {
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
        originalInput: 'name: awesome-batch-changes\ndescription: somestring',
        supersedingBatchSpec: null,
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
