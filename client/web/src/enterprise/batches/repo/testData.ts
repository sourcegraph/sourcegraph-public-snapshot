import { subDays } from 'date-fns'

import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
} from '@sourcegraph/shared/src/graphql-operations'

import { ChangesetFields, RepoBatchChange } from '../../../graphql-operations'

export const now = new Date()

const READY_EXTERNAL_CHANGESET: ChangesetFields = {
    __typename: 'ExternalChangeset',
    body: 'body',
    checkState: ChangesetCheckState.PASSED,
    diffStat: {
        added: 10,
        changed: 9,
        deleted: 1,
    },
    externalID: '123',
    externalURL: {
        url: 'http://test.test/123',
    },
    labels: [{ color: '93ba13', description: 'Very awesome description', text: 'Some label' }],
    repository: {
        id: 'repoid',
        name: 'github.com/sourcegraph/awesome',
        url: 'http://test.test/awesome',
    },
    reviewState: ChangesetReviewState.COMMENTED,
    title: 'Add prettier to all projects',
    createdAt: subDays(now, 10).toISOString(),
    updatedAt: subDays(now, 1).toISOString(),
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
            headRef: 'my-branch',
        },
    },
}

const FAILED_EXTERNAL_CHANGESET: ChangesetFields = {
    __typename: 'ExternalChangeset',
    body: 'body',
    checkState: null,
    diffStat: {
        added: 10,
        changed: 9,
        deleted: 1,
    },
    externalID: null,
    externalURL: null,
    labels: [],
    repository: {
        id: 'repoid',
        name: 'github.com/sourcegraph/awesome',
        url: 'http://test.test/awesome',
    },
    reviewState: null,
    title: 'Add prettier to all projects',
    createdAt: subDays(now, 30).toISOString(),
    updatedAt: subDays(now, 2).toISOString(),
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
            headRef: 'my-branch',
        },
    },
}

export const NODES: RepoBatchChange[] = [
    {
        id: 'test',
        url: '/users/alice/batch-change/test',
        name: 'Awesome batch',
        description: `# What this does

This is my thorough explanation. And it can also get very long, in that case the UI doesn't break though, which is good. And one more line to finally be longer than the viewport.`,
        createdAt: subDays(now, 1).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        changesets: {
            totalCount: 2,
            pageInfo: { endCursor: null, hasNextPage: false },
            nodes: [
                READY_EXTERNAL_CHANGESET,
                READY_EXTERNAL_CHANGESET,
                READY_EXTERNAL_CHANGESET,
                READY_EXTERNAL_CHANGESET,
            ],
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
    {
        id: 'test2',
        url: '/users/alice/batch-changes/test2',
        name: 'Awesome batch',
        description: null,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        changesets: {
            totalCount: 1,
            pageInfo: { endCursor: null, hasNextPage: false },
            nodes: [READY_EXTERNAL_CHANGESET],
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
    {
        id: 'test3',
        url: '/users/alice/batch-changes/test3',
        name: 'Awesome batch',
        description: `# My batch

        This is my thorough explanation.`,
        createdAt: subDays(now, 30).toISOString(),
        closedAt: subDays(now, 3).toISOString(),
        changesetsStats: {
            open: 0,
            closed: 10,
            merged: 5,
        },
        changesets: {
            totalCount: 2,
            pageInfo: { endCursor: null, hasNextPage: false },
            nodes: [FAILED_EXTERNAL_CHANGESET],
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
]
