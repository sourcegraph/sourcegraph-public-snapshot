import { boolean } from '@storybook/addon-knobs'
import { useMemo, useCallback } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
    BatchChangeFields,
} from '../../../graphql-operations'
import {
    queryChangesets as _queryChangesets,
    queryExternalChangesetWithFileDiffs,
    fetchBatchChangeByNamespace,
} from '../detail/backend'

import { BatchChangeClosePage } from './BatchChangeClosePage'

const { add } = storiesOf('web/batches/close/BatchChangeClosePage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

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
        total: 10,
        archived: 18,
        unpublished: 4,
    },
    createdAt: subDays(now, 5).toISOString(),
    initialApplier: {
        url: '/users/alice',
        username: 'alice',
    },
    id: 'specid',
    url: '/users/alice/batch-changes/specid',
    namespace: {
        namespaceName: 'alice',
        url: '/users/alice',
    },
    diffStat: { added: 1000, changed: 2000, deleted: 1000, __typename: 'DiffStat' },
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
        originalInput: 'name: awesome-batch-change\ndescription: somestring',
        supersedingBatchSpec: null,
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
}

const queryChangesets: typeof _queryChangesets = () =>
    of({
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
                        headRef: 'my-branch',
                    },
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
                        headRef: 'my-branch',
                    },
                },
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

add('Overview', () => {
    const viewerCanAdminister = boolean('viewerCanAdminister', true)
    const batchChange: BatchChangeFields = useMemo(
        () => ({
            ...batchChangeDefaults,
            viewerCanAdminister,
        }),
        [viewerCanAdminister]
    )
    const fetchBatchChange: typeof fetchBatchChangeByNamespace = useCallback(() => of(batchChange), [batchChange])
    return (
        <WebStory>
            {props => (
                <BatchChangeClosePage
                    {...props}
                    queryChangesets={queryChangesets}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    namespaceID="n123"
                    batchChangeName="c123"
                    fetchBatchChangeByNamespace={fetchBatchChange}
                    extensionsController={{} as any}
                    platformContext={{} as any}
                />
            )}
        </WebStory>
    )
})

add('No open changesets', () => {
    const batchChange: BatchChangeFields = useMemo(() => batchChangeDefaults, [])
    const fetchBatchChange: typeof fetchBatchChangeByNamespace = useCallback(() => of(batchChange), [batchChange])
    const queryEmptyChangesets = useCallback(
        () =>
            of({
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
        <WebStory>
            {props => (
                <BatchChangeClosePage
                    {...props}
                    queryChangesets={queryEmptyChangesets}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    namespaceID="n123"
                    batchChangeName="c123"
                    fetchBatchChangeByNamespace={fetchBatchChange}
                    extensionsController={{} as any}
                    platformContext={{} as any}
                />
            )}
        </WebStory>
    )
})
