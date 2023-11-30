import type { StoryFn, Meta, Decorator } from '@storybook/react'
import classNames from 'classnames'
import { addHours } from 'date-fns'
import { of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'
import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
} from '../../../../graphql-operations'

import { ExternalChangesetNode } from './ExternalChangesetNode'

import gridStyles from './BatchChangeChangesets.module.scss'

const decorator: Decorator = story => (
    <div className={classNames(gridStyles.batchChangeChangesetsGrid, 'p-3 container')}>{story()}</div>
)

const config: Meta = {
    title: 'web/batches/ExternalChangesetNode',
    decorators: [decorator],
    argTypes: {
        viewerCanAdminister: {
            control: { type: 'boolean' },
        },
        labeled: {
            control: { type: 'boolean' },
        },
        commitsSigned: {
            control: { type: 'boolean' },
        },
    },
    args: {
        viewerCanAdminister: true,
        labeled: true,
        commitsSigned: true,
    },
}

export default config

export const AllStates: StoryFn = args => {
    const now = new Date()
    return (
        <WebStory>
            {props => (
                <>
                    {Object.values(ChangesetState)
                        .filter(state => state !== ChangesetState.UNPUBLISHED)
                        .map((state, index) => (
                            <ExternalChangesetNode
                                key={index}
                                {...props}
                                node={{
                                    id: 'somechangeset',
                                    updatedAt: now.toISOString(),
                                    nextSyncAt: addHours(now, 1).toISOString(),
                                    state,
                                    __typename: 'ExternalChangeset',
                                    title: 'Changeset title on code host',
                                    error: null,
                                    syncerError: null,
                                    body: 'This changeset does the following things:\nIs awesome\nIs useful',
                                    checkState: ChangesetCheckState.PENDING,
                                    createdAt: now.toISOString(),
                                    commitVerification: args.commitsSigned ? { verified: true } : null,
                                    externalID: '123',
                                    externalURL: {
                                        url: 'http://test.test/pr/123',
                                    },
                                    forkNamespace: index % 2 === 0 ? 'user' : null,
                                    diffStat: {
                                        __typename: 'DiffStat',
                                        added: 30,
                                        deleted: 28,
                                    },
                                    labels: args.labeled
                                        ? [
                                              {
                                                  __typename: 'ChangesetLabel',
                                                  color: '93ba13',
                                                  description: 'Very awesome description',
                                                  text: 'Some label',
                                              },
                                          ]
                                        : [],
                                    repository: {
                                        id: 'repoid',
                                        name: 'github.com/sourcegraph/sourcegraph',
                                        url: 'http://test.test/sourcegraph/sourcegraph',
                                    },
                                    reviewState: ChangesetReviewState.COMMENTED,
                                    currentSpec: {
                                        id: 'spec-rand-id-1',
                                        type: ChangesetSpecType.BRANCH,
                                        description: {
                                            __typename: 'GitBranchChangesetDescription',
                                            baseRef: 'my-branch',
                                            headRef: 'my-branch',
                                        },
                                        forkTarget:
                                            index % 2 === 0
                                                ? {
                                                      pushUser: true,
                                                      namespace: null,
                                                  }
                                                : { pushUser: false, namespace: null },
                                    },
                                }}
                                viewerCanAdminister={args.viewerCanAdminister}
                                queryExternalChangesetWithFileDiffs={() =>
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
                                }
                            />
                        ))}
                </>
            )}
        </WebStory>
    )
}

AllStates.storyName = 'All states'

export const Unpublished: StoryFn = args => {
    const now = new Date()
    return (
        <WebStory>
            {props => (
                <ExternalChangesetNode
                    {...props}
                    node={{
                        __typename: 'ExternalChangeset',
                        id: 'somechangeset',
                        updatedAt: now.toISOString(),
                        nextSyncAt: null,
                        state: ChangesetState.UNPUBLISHED,
                        title: 'Changeset title on code host',
                        error: null,
                        syncerError: null,
                        body: 'This changeset does the following things:\nIs awesome\nIs useful',
                        checkState: null,
                        createdAt: now.toISOString(),
                        commitVerification: null,
                        externalID: null,
                        externalURL: null,
                        forkNamespace: null,
                        diffStat: {
                            __typename: 'DiffStat',
                            added: 30,
                            deleted: 28,
                        },
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
                            name: 'github.com/sourcegraph/sourcegraph',
                            url: 'http://test.test/sourcegraph/sourcegraph',
                        },
                        reviewState: null,
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
                    }}
                    viewerCanAdminister={args.viewerCanAdminister}
                    queryExternalChangesetWithFileDiffs={() =>
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
                    }
                />
            )}
        </WebStory>
    )
}

export const Importing: StoryFn = args => {
    const now = new Date()
    return (
        <WebStory>
            {props => (
                <ExternalChangesetNode
                    {...props}
                    node={{
                        __typename: 'ExternalChangeset',
                        id: 'somechangeset',
                        updatedAt: now.toISOString(),
                        nextSyncAt: null,
                        state: ChangesetState.PROCESSING,
                        // No title yet, still importing.
                        title: null,
                        error: null,
                        syncerError: null,
                        body: null,
                        checkState: null,
                        createdAt: now.toISOString(),
                        commitVerification: null,
                        externalID: '12345',
                        externalURL: null,
                        forkNamespace: null,
                        diffStat: null,
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
                            name: 'github.com/sourcegraph/sourcegraph',
                            url: 'http://test.test/sourcegraph/sourcegraph',
                        },
                        reviewState: null,
                        currentSpec: null,
                    }}
                    viewerCanAdminister={args.viewerCanAdminister}
                    queryExternalChangesetWithFileDiffs={() =>
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
                    }
                />
            )}
        </WebStory>
    )
}

export const ImportingFailed: StoryFn = args => {
    const now = new Date()
    return (
        <WebStory>
            {props => (
                <ExternalChangesetNode
                    {...props}
                    node={{
                        __typename: 'ExternalChangeset',
                        id: 'somechangeset-2',
                        updatedAt: now.toISOString(),
                        nextSyncAt: null,
                        state: ChangesetState.FAILED,
                        // No title, because it wasn't found.
                        title: null,
                        error: 'Changeset with external ID 99999 not found',
                        syncerError: null,
                        body: null,
                        checkState: null,
                        createdAt: now.toISOString(),
                        commitVerification: null,
                        externalID: '99999',
                        externalURL: null,
                        forkNamespace: null,
                        diffStat: null,
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
                            name: 'github.com/sourcegraph/sourcegraph',
                            url: 'http://test.test/sourcegraph/sourcegraph',
                        },
                        reviewState: null,
                        currentSpec: null,
                    }}
                    viewerCanAdminister={args.viewerCanAdminister}
                    queryExternalChangesetWithFileDiffs={() =>
                        of({
                            diff: null,
                        })
                    }
                />
            )}
        </WebStory>
    )
}

ImportingFailed.storyName = 'Importing failed'

export const SyncFailed: StoryFn = args => {
    const now = new Date()
    return (
        <WebStory>
            {props => (
                <ExternalChangesetNode
                    {...props}
                    node={{
                        __typename: 'ExternalChangeset',
                        id: 'somechangeset-2',
                        updatedAt: now.toISOString(),
                        nextSyncAt: null,
                        state: ChangesetState.FAILED,
                        // No title, because it wasn't found.
                        title: null,
                        error: null,
                        syncerError: 'Invalid token, cannot load PR.',
                        body: null,
                        checkState: null,
                        createdAt: now.toISOString(),
                        commitVerification: null,
                        externalID: '99999',
                        externalURL: null,
                        forkNamespace: null,
                        diffStat: null,
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
                            name: 'github.com/sourcegraph/sourcegraph',
                            url: 'http://test.test/sourcegraph/sourcegraph',
                        },
                        reviewState: null,
                        currentSpec: null,
                    }}
                    viewerCanAdminister={args.viewerCanAdminister}
                    queryExternalChangesetWithFileDiffs={() =>
                        of({
                            diff: null,
                        })
                    }
                />
            )}
        </WebStory>
    )
}

SyncFailed.storyName = 'Sync failed'
