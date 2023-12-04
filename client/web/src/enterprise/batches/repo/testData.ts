import { subDays } from 'date-fns'

import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
} from '@sourcegraph/shared/src/graphql-operations'

import type { ChangesetFields, RepoBatchChange } from '../../../graphql-operations'

export const now = new Date()

const READY_EXTERNAL_CHANGESET: ChangesetFields = {
    __typename: 'ExternalChangeset',
    body: 'body',
    checkState: ChangesetCheckState.PASSED,
    diffStat: {
        __typename: 'DiffStat',
        added: 19,
        deleted: 10,
    },
    externalID: '123',
    externalURL: {
        url: 'http://test.test/123',
    },
    forkNamespace: null,
    labels: [
        { __typename: 'ChangesetLabel', color: '93ba13', description: 'Very awesome description', text: 'Some label' },
    ],
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
    commitVerification: null,
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
}

const FAILED_EXTERNAL_CHANGESET: ChangesetFields = {
    __typename: 'ExternalChangeset',
    body: 'body',
    checkState: null,
    diffStat: {
        __typename: 'DiffStat',
        added: 19,
        deleted: 10,
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
    createdAt: subDays(now, 30).toISOString(),
    updatedAt: subDays(now, 2).toISOString(),
    state: ChangesetState.RETRYING,
    nextSyncAt: null,
    id: 'somev2',
    error: 'Cannot create PR, insufficient token scope.',
    syncerError: null,
    commitVerification: null,
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
}

const commonFields = (id: number) =>
    ({
        id: `test${id}`,
        url: `/users/alice/batch-change/test${id}`,
        name: `Awesome batch ${id}`,
        description: `# My batch

    This is my thorough explanation.`,
        changesetsStats: {
            open: 0,
            closed: 10,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    } as const)

export const NODES: RepoBatchChange[] = [
    {
        ...commonFields(1),
        description: `# What this does

This is my thorough explanation. And it can also get very long, in that case the UI doesn't break though, which is good. And one more line to finally be longer than the viewport.`,
        createdAt: subDays(now, 1).toISOString(),
        closedAt: null,
        changesets: {
            totalCount: 25,
            pageInfo: { endCursor: null, hasNextPage: false },
            nodes: (new Array(10) as typeof READY_EXTERNAL_CHANGESET[]).fill(READY_EXTERNAL_CHANGESET),
        },
    },
    {
        ...commonFields(2),
        description: null,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesets: {
            totalCount: 1,
            pageInfo: { endCursor: null, hasNextPage: false },
            nodes: [READY_EXTERNAL_CHANGESET],
        },
    },
    {
        ...commonFields(3),
        createdAt: subDays(now, 30).toISOString(),
        closedAt: subDays(now, 3).toISOString(),
        changesets: {
            totalCount: 2,
            pageInfo: { endCursor: null, hasNextPage: false },
            nodes: [FAILED_EXTERNAL_CHANGESET, READY_EXTERNAL_CHANGESET],
        },
    },
]
