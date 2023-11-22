import { subDays } from 'date-fns'

import {
    type BatchChangesByNamespaceResult,
    type BatchChangesResult,
    BatchChangeState,
    BatchSpecState,
    type GetLicenseAndUsageInfoResult,
    type ListBatchChange,
} from '../../../graphql-operations'

export const now = new Date()

export const nodes: Record<string, { __typename: 'BatchChange' } & ListBatchChange> = {
    'Open batch change': {
        __typename: 'BatchChange',
        id: 'test',
        url: '/users/alice/batch-change/test',
        name: 'Awesome batch',
        state: BatchChangeState.OPEN,
        description: `# What this does

This is my thorough explanation. And it can also get very long, in that case the UI doesn't break though, which is good. And one more line to finally be longer than the viewport.`,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            __typename: 'BatchSpec',
            id: 'old-spec-1',
            state: BatchSpecState.COMPLETED,
            applyURL: null,
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-1',
                    state: BatchSpecState.PROCESSING,
                    applyURL: null,
                },
            ],
        },
    },
    'Failed draft': {
        __typename: 'BatchChange',
        id: 'testdraft',
        url: '/users/alice/batch-change/test',
        name: 'Awesome batch',
        state: BatchChangeState.DRAFT,
        description: 'The execution of the batch spec failed.',
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            __typename: 'BatchSpec',
            id: 'empty-draft',
            state: BatchSpecState.PENDING,
            applyURL: null,
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-2',
                    state: BatchSpecState.FAILED,
                    applyURL: null,
                },
            ],
        },
    },
    'No description': {
        __typename: 'BatchChange',
        id: 'test2',
        url: '/users/alice/batch-changes/test2',
        name: 'Awesome batch',
        state: BatchChangeState.OPEN,
        description: null,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            __typename: 'BatchSpec',
            id: 'empty-draft',
            state: BatchSpecState.PENDING,
            applyURL: null,
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-3',
                    state: BatchSpecState.COMPLETED,
                    applyURL: '/fake-apply-url',
                },
            ],
        },
    },
    'Closed batch change': {
        __typename: 'BatchChange',
        id: 'test3',
        url: '/users/alice/batch-changes/test3',
        name: 'Awesome batch',
        state: BatchChangeState.CLOSED,
        description: `# My batch

        This is my thorough explanation.`,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: subDays(now, 3).toISOString(),
        changesetsStats: {
            open: 0,
            closed: 10,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            __typename: 'BatchSpec',
            id: 'empty-draft',
            state: BatchSpecState.PENDING,
            applyURL: null,
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-4',
                    state: BatchSpecState.COMPLETED,
                    applyURL: '/fake-apply-url',
                },
            ],
        },
    },
}

export const BATCH_CHANGES_RESULT: BatchChangesResult = {
    batchChanges: {
        __typename: 'BatchChangeConnection',
        totalCount: Object.values(nodes).length,
        nodes: Object.values(nodes),
        pageInfo: { endCursor: null, hasNextPage: false },
    },
}

export const NO_BATCH_CHANGES_RESULT: BatchChangesResult = {
    batchChanges: {
        __typename: 'BatchChangeConnection',
        totalCount: 0,
        nodes: [],
        pageInfo: { endCursor: null, hasNextPage: false },
    },
}

export const BATCH_CHANGES_BY_NAMESPACE_RESULT: BatchChangesByNamespaceResult = {
    node: {
        __typename: 'User',
        batchChanges: BATCH_CHANGES_RESULT.batchChanges,
    },
}

export const getLicenseAndUsageInfoResult = (
    isLicensed = true,
    hasBatchChanges = true,
    maxUnlicensedChangesets = 10
): GetLicenseAndUsageInfoResult => ({
    campaigns: isLicensed,
    batchChanges: isLicensed,
    allBatchChanges: { totalCount: hasBatchChanges ? Object.values(nodes).length : 0 },
    maxUnlicensedChangesets,
})
