import type { Meta, Decorator, StoryFn } from '@storybook/react'
import { WildcardMockLink, MATCH_ANY_PARAMETERS } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import type { GlobalChangesetsStatsResult } from '../../../graphql-operations'

import { GLOBAL_CHANGESETS_STATS } from './backend'
import { BatchChangeStatsBar } from './BatchChangeStatsBar'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/list/BatchChangeStatsBar',
    decorators: [decorator],
}

export default config

const statBarData: GlobalChangesetsStatsResult = {
    __typename: 'Query',
    batchChanges: { __typename: 'BatchChangeConnection', totalCount: 30 },
    globalChangesetsStats: { __typename: 'GlobalChangesetsStats', open: 7, closed: 5, merged: 21 },
}

export const BatchChangeStatsBarStory: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(GLOBAL_CHANGESETS_STATS),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: statBarData,
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <BatchChangeStatsBar {...props} className="text-center" />
            </MockedTestProvider>
        )}
    </WebStory>
)

BatchChangeStatsBarStory.storyName = 'BatchChangeStatsBar'
