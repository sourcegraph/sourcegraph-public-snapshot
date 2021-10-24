import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import { addHours } from 'date-fns'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'
import {
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
} from '../../../../graphql-operations'

import gridStyles from './BatchChangeChangesets.module.scss'
import { ExternalChangesetNode } from './ExternalChangesetNode'

const { add } = storiesOf('web/batches/ExternalChangesetNode', module).addDecorator(story => (
    <div className={classNames(gridStyles.batchChangeChangesetsGrid, 'p-3 container')}>{story()}</div>
))

add('All states', () => {
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
                                    externalID: '123',
                                    externalURL: {
                                        url: 'http://test.test/pr/123',
                                    },
                                    diffStat: {
                                        __typename: 'DiffStat',
                                        added: 10,
                                        changed: 20,
                                        deleted: 8,
                                    },
                                    labels: [
                                        {
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
                                    reviewState: ChangesetReviewState.COMMENTED,
                                    currentSpec: {
                                        id: 'spec-rand-id-1',
                                        type: ChangesetSpecType.BRANCH,
                                        description: {
                                            __typename: 'GitBranchChangesetDescription',
                                            headRef: 'my-branch',
                                        },
                                    },
                                }}
                                viewerCanAdminister={boolean('viewerCanAdminister', true)}
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
})

add('Unpublished', () => {
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
                        externalID: null,
                        externalURL: null,
                        diffStat: {
                            __typename: 'DiffStat',
                            added: 10,
                            changed: 20,
                            deleted: 8,
                        },
                        labels: [{ color: '93ba13', description: 'Very awesome description', text: 'Some label' }],
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
                                headRef: 'my-branch',
                            },
                        },
                    }}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
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
})

add('Importing', () => {
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
                        externalID: '12345',
                        externalURL: null,
                        diffStat: null,
                        labels: [{ color: '93ba13', description: 'Very awesome description', text: 'Some label' }],
                        repository: {
                            id: 'repoid',
                            name: 'github.com/sourcegraph/sourcegraph',
                            url: 'http://test.test/sourcegraph/sourcegraph',
                        },
                        reviewState: null,
                        currentSpec: null,
                    }}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
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
})

add('Importing failed', () => {
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
                        externalID: '99999',
                        externalURL: null,
                        diffStat: null,
                        labels: [{ color: '93ba13', description: 'Very awesome description', text: 'Some label' }],
                        repository: {
                            id: 'repoid',
                            name: 'github.com/sourcegraph/sourcegraph',
                            url: 'http://test.test/sourcegraph/sourcegraph',
                        },
                        reviewState: null,
                        currentSpec: null,
                    }}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    queryExternalChangesetWithFileDiffs={() =>
                        of({
                            diff: null,
                        })
                    }
                />
            )}
        </WebStory>
    )
})

add('Sync failed', () => {
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
                        externalID: '99999',
                        externalURL: null,
                        diffStat: null,
                        labels: [{ color: '93ba13', description: 'Very awesome description', text: 'Some label' }],
                        repository: {
                            id: 'repoid',
                            name: 'github.com/sourcegraph/sourcegraph',
                            url: 'http://test.test/sourcegraph/sourcegraph',
                        },
                        reviewState: null,
                        currentSpec: null,
                    }}
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    queryExternalChangesetWithFileDiffs={() =>
                        of({
                            diff: null,
                        })
                    }
                />
            )}
        </WebStory>
    )
})
