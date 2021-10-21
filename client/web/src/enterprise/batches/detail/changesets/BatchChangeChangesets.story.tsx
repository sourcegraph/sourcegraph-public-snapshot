import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { addHours } from 'date-fns'
import { noop } from 'lodash'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../../components/WebStory'
import {
    ChangesetFields,
    ChangesetCheckState,
    ChangesetReviewState,
    ChangesetSpecType,
    ChangesetState,
} from '../../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs } from '../backend'

import { BatchChangeChangesets } from './BatchChangeChangesets'

const { add } = storiesOf('web/batches/BatchChangeChangesets', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const now = new Date()
const nodes: ChangesetFields[] = [
    ...Object.values(ChangesetState).map(
        (state): ChangesetFields => ({
            __typename: 'ExternalChangeset',
            id: 'somechangeset' + state,
            updatedAt: now.toISOString(),
            nextSyncAt: addHours(now, 1).toISOString(),
            state,
            title: 'Changeset title on code host',
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
                    headRef: 'my-branch',
                },
            },
        })
    ),
    ...Object.values(ChangesetState).map(
        (state): ChangesetFields => ({
            __typename: 'HiddenExternalChangeset' as const,
            id: 'somehiddenchangeset' + state,
            updatedAt: now.toISOString(),
            nextSyncAt: addHours(now, 1).toISOString(),
            state,
            createdAt: now.toISOString(),
        })
    ),
]
const queryChangesets = () => of({ totalCount: nodes.length, nodes, pageInfo: { endCursor: null, hasNextPage: false } })

const queryEmptyExternalChangesetWithFileDiffs: typeof queryExternalChangesetWithFileDiffs = ({
    externalChangeset,
}) => {
    switch (externalChangeset) {
        case 'somechangesetCLOSED':
        case 'somechangesetMERGED':
        case 'somechangesetDELETED':
            return of({
                diff: null,
            })
        default:
            return of({
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
}

add('List of changesets', () => (
    <WebStory>
        {props => (
            <BatchChangeChangesets
                {...props}
                refetchBatchChange={noop}
                queryChangesets={queryChangesets}
                queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                extensionsController={undefined as any}
                platformContext={undefined as any}
                batchChangeID="batchid"
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
            />
        )}
    </WebStory>
))

add('List of expanded changesets', () => (
    <WebStory>
        {props => (
            <BatchChangeChangesets
                {...props}
                refetchBatchChange={noop}
                queryChangesets={queryChangesets}
                queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                extensionsController={undefined as any}
                platformContext={undefined as any}
                batchChangeID="batchid"
                viewerCanAdminister={boolean('viewerCanAdminister', true)}
                expandByDefault={true}
            />
        )}
    </WebStory>
))
