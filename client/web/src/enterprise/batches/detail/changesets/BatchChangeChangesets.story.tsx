import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { of } from 'rxjs'
import { WildcardMockLink, MATCH_ANY_PARAMETERS } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { CHANGESETS, queryExternalChangesetWithFileDiffs } from '../backend'

import { BatchChangeChangesets } from './BatchChangeChangesets'
import { BATCH_CHANGE_CHANGESETS_RESULT } from './BatchChangeChangesets.mock'

const { add } = storiesOf('web/batches/BatchChangeChangesets', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const mocks = new WildcardMockLink([
    {
        request: { query: getDocumentNode(CHANGESETS), variables: MATCH_ANY_PARAMETERS },
        result: { data: { node: BATCH_CHANGE_CHANGESETS_RESULT } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

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
            <MockedTestProvider link={mocks}>
                <BatchChangeChangesets
                    {...props}
                    refetchBatchChange={noop}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    extensionsController={undefined as any}
                    platformContext={undefined as any}
                    batchChangeID="batchid"
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('List of expanded changesets', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={mocks}>
                <BatchChangeChangesets
                    {...props}
                    refetchBatchChange={noop}
                    queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                    extensionsController={undefined as any}
                    platformContext={undefined as any}
                    batchChangeID="batchid"
                    viewerCanAdminister={boolean('viewerCanAdminister', true)}
                    expandByDefault={true}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))
