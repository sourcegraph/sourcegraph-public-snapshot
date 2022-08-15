import { boolean, select } from '@storybook/addon-knobs'
import { Story, Meta, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'
import { of } from 'rxjs'
import { WildcardMockLink, MATCH_ANY_PARAMETERS } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { BatchChangeState } from '../../../../graphql-operations'
import { CHANGESETS, queryExternalChangesetWithFileDiffs } from '../backend'

import { BatchChangeChangesets } from './BatchChangeChangesets'
import { BATCH_CHANGE_CHANGESETS_RESULT, EMPTY_BATCH_CHANGE_CHANGESETS_RESULT } from './BatchChangeChangesets.mock'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangeChangesets',
    decorators: [decorator],
}

export default config

const mocks = new WildcardMockLink([
    {
        request: { query: getDocumentNode(CHANGESETS), variables: MATCH_ANY_PARAMETERS },
        result: { data: { node: BATCH_CHANGE_CHANGESETS_RESULT } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

const emptyMockData = new WildcardMockLink([
    {
        request: { query: getDocumentNode(CHANGESETS), variables: MATCH_ANY_PARAMETERS },
        result: { data: { node: EMPTY_BATCH_CHANGE_CHANGESETS_RESULT } },
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

export const ListOfChangesets: Story = () => (
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
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                    batchChangeState={BatchChangeState.OPEN}
                    isExecutionEnabled={false}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ListOfChangesets.storyName = 'List of changesets'

export const ListOfExpandedChangesets: Story = () => (
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
                    settingsCascade={EMPTY_SETTINGS_CASCADE}
                    batchChangeState={BatchChangeState.OPEN}
                    isExecutionEnabled={false}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

ListOfExpandedChangesets.storyName = 'List of expanded changesets'

export const DraftWithoutChangesets: Story = () => {
    const options = Object.keys(BatchChangeState)
    const batchChangeState = select('batchChangeState', options, BatchChangeState.DRAFT)

    return (
        <WebStory>
            {props => (
                <MockedTestProvider link={emptyMockData}>
                    <BatchChangeChangesets
                        {...props}
                        refetchBatchChange={noop}
                        queryExternalChangesetWithFileDiffs={queryEmptyExternalChangesetWithFileDiffs}
                        extensionsController={undefined as any}
                        platformContext={undefined as any}
                        batchChangeID="batchid"
                        viewerCanAdminister={true}
                        expandByDefault={true}
                        settingsCascade={EMPTY_SETTINGS_CASCADE}
                        batchChangeState={batchChangeState as BatchChangeState}
                        isExecutionEnabled={boolean('isExecutionEnabled', true)}
                    />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

DraftWithoutChangesets.storyName = 'Draft without changesets'
