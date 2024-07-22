import { useMemo, useCallback } from '@storybook/addons'
import type { Meta, StoryFn, Decorator } from '@storybook/react'
import { subDays } from 'date-fns'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
    type BatchChangeFields,
    BatchSpecState,
    BatchChangeState,
    BatchSpecSource,
} from '../../../graphql-operations'
import type {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs,
    fetchBatchChangeByNamespace,
} from '../detail/backend'

import { BatchChangeClosePage } from './BatchChangeClosePage'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/close/BatchChangeClosePage',
    decorators: [decorator],
    parameters: {},
}

export default config

const now = new Date()

const batchChangeDefaults: BatchChangeFields = {
    __typename: 'BatchChange',
    changesetsStats: {
        __typename: 'ChangesetsStats',
        closed: 1,
        deleted: 1,
        merged: 2,
        draft: 1,
        open: 2,
        total: 29,
        archived: 18,
        unpublished: 4,
        isCompleted: false,
        percentComplete: 30,
        failed: 0,
        retrying: 0,
        scheduled: 0,
        processing: 0,
    },
    createdAt: subDays(now, 5).toISOString(),
    creator: {
        url: '/users/alice',
        username: 'alice',
    },
    id: 'specid',
    url: '/users/alice/batch-changes/specid',
    namespace: {
        __typename: 'User',
        id: '1234',
        displayName: null,
        username: 'alice',
        namespaceName: 'alice',
        url: '/users/alice',
    },
    diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
    viewerCanAdminister: true,
    closedAt: null,
    description: '## What this batch change does\n\nTruly awesome things for example.',
    name: 'awesome-batch-change',
    updatedAt: subDays(now, 5).toISOString(),
    lastAppliedAt: subDays(now, 5).toISOString(),
    lastApplier: {
        url: '/users/bob',
        username: 'bob',
    },
    currentSpec: {
        id: 'specID1',
        originalInput: 'name: awesome-batch-change\ndescription: somestring',
        supersedingBatchSpec: null,
        source: BatchSpecSource.REMOTE,
        codeHostsWithoutWebhooks: {
            nodes: [],
            pageInfo: { hasNextPage: false },
            totalCount: 0,
        },
        viewerBatchChangesCodeHosts: {
            __typename: 'BatchChangesCodeHostConnection',
            totalCount: 0,
            nodes: [],
        },
        files: null,
        description: {
            __typename: 'BatchChangeDescription',
            name: 'Spec Description',
        },
    },
    batchSpecs: {
        nodes: [{ state: BatchSpecState.COMPLETED }],
        pageInfo: { hasNextPage: false },
    },
    bulkOperations: {
        __typename: 'BulkOperationConnection',
        totalCount: 0,
    },
    activeBulkOperations: {
        __typename: 'BulkOperationConnection',
        totalCount: 0,
        nodes: [],
    },
    state: BatchChangeState.OPEN,
}

const queryChangesets: typeof _queryChangesets = () =>
    of({
        __typename: 'ChangesetConnection',
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: 6,
        nodes: [
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.OPEN,
                id: 'someh1',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.OPEN,
                id: 'someh2',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.OPEN,
                id: 'someh3',
                nextSyncAt: null,
                updatedAt: subDays(now, 5).toISOString(),
            },
            {
                __typename: 'HiddenExternalChangeset',
                createdAt: subDays(now, 5).toISOString(),
                state: ChangesetState.OPEN,
                id: 'someh4',
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
                commitVerification: null,
            },
            {
                __typename: 'ExternalChangeset',
                body: 'body',
                checkState: null,
                diffStat: {
                    __typename: 'DiffStat',
                    added: 10,
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
                state: ChangesetState.OPEN,
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
                commitVerification: null,
            },
        ],
    })

const queryEmptyExternalChangesetWithFileDiffs: typeof queryExternalChangesetWithFileDiffs = () =>
    of({
        diff: {
            __typename: 'PreviewRepositoryComparison',
            fileDiffs: {
                nodes: [],
                totalCount: 0,
                pageInfo: {
                    endCursor: null,
                    hasNextPage: false,
                },
            },
        },
    })

export const Overview: StoryFn = args => {
    const viewerCanAdminister = args.viewerCanAdminister
    const batchChange: BatchChangeFields = useMemo(
        () => ({
            ...batchChangeDefaults,
            viewerCanAdminister,
        }),
        [viewerCanAdminister]
    )
    const fetchBatchChange: typeof fetchBatchChangeByNamespace = useCallback(() => of(batchChange), [batchChange])
    return (
        <WebStory path="/:batchChangeName" initialEntries={['/c123']}>
            {props => (
                <BatchChangeClosePage
                    {...props}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    namespaceID="n123"
                    fetchBatchChangeByNamespace={fetchBatchChange}
                />
            )}
        </WebStory>
    )
}
Overview.argTypes = {
    viewerCanAdminister: {
        control: { type: 'boolean' },
    },
}
Overview.args = {
    viewerCanAdminister: true,
}

export const NoOpenChangesets: StoryFn = () => {
    const batchChange: BatchChangeFields = useMemo(() => batchChangeDefaults, [])
    const fetchBatchChange: typeof fetchBatchChangeByNamespace = useCallback(() => of(batchChange), [batchChange])
    const queryEmptyChangesets = useCallback(
        () =>
            of({
                __typename: 'ChangesetConnection' as const,
                pageInfo: {
                    endCursor: null,
                    hasNextPage: false,
                },
                totalCount: 0,
                nodes: [],
            }),
        []
    )
    return (
        <WebStory path="/:batchChangeName" initialEntries={['/c123']}>
            {props => (
                <BatchChangeClosePage
                    {...props}
                    queryChangesets={queryEmptyChangesets}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    namespaceID="n123"
                    fetchBatchChangeByNamespace={fetchBatchChange}
                />
            )}
        </WebStory>
    )
}

NoOpenChangesets.storyName = 'No open changesets'
