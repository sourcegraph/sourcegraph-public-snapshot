import { addHours } from 'date-fns'

import {
    type ExternalChangesetFields,
    type HiddenExternalChangesetFields,
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
    type BatchChangeChangesetsResult,
} from '../../../../graphql-operations'
import { MOCK_BATCH_CHANGE } from '../BatchChangeDetailsPage.mock'

const now = new Date()

export const BATCH_CHANGE_CHANGESETS: (ExternalChangesetFields | HiddenExternalChangesetFields)[] = [
    ...Object.values(ChangesetState).map(
        (state): ExternalChangesetFields => ({
            __typename: 'ExternalChangeset',
            id: 'somechangeset' + state,
            updatedAt: now.toISOString(),
            nextSyncAt: addHours(now, 1).toISOString(),
            state,
            title: 'Changeset title on code host',
            body: 'This changeset does the following things:\nIs awesome\nIs useful',
            checkState: ChangesetCheckState.PENDING,
            createdAt: now.toISOString(),
            commitVerification: null,
            externalID: '123',
            externalURL: {
                url: 'http://test.test/pr/123',
            },
            forkNamespace: null,
            diffStat: {
                __typename: 'DiffStat',
                added: 30,
                deleted: 28,
            },
            labels: [],
            repository: {
                id: 'repoid',
                name: 'github.com/sourcegraph/sourcegraph',
                url: 'http://test.test/sourcegraph/sourcegraph',
            },
            reviewState: ChangesetReviewState.COMMENTED,
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
        })
    ),
    ...Object.values(ChangesetState).map(
        (state): HiddenExternalChangesetFields => ({
            __typename: 'HiddenExternalChangeset' as const,
            id: 'somehiddenchangeset' + state,
            updatedAt: now.toISOString(),
            nextSyncAt: addHours(now, 1).toISOString(),
            state,
            createdAt: now.toISOString(),
        })
    ),
]

export const BATCH_CHANGE_CHANGESETS_RESULT: BatchChangeChangesetsResult['node'] = {
    ...MOCK_BATCH_CHANGE,
    changesets: {
        __typename: 'ChangesetConnection',
        totalCount: BATCH_CHANGE_CHANGESETS.length,
        nodes: BATCH_CHANGE_CHANGESETS,
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
