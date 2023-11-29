import type { MockedResponse } from '@apollo/client/testing'
import type { Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { GroupByField, TimeIntervalStepUnit } from '@sourcegraph/shared/src/graphql-operations'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../../../../components/WebStory'
import type { GetInsightPreviewResult } from '../../../../../../../graphql-operations'
import type { SearchBasedInsightSeries } from '../../../../../core'
import { GET_INSIGHT_PREVIEW_GQL } from '../../../../../core/hooks/live-preview-insight'

import { ComputeLivePreview as ComputeLivePreviewComponent } from './ComputeLivePreview'

const defaultStory: Meta = {
    title: 'web/insights/creation-ui/compute/ComputeLivePreview',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

const mock: MockedResponse<GetInsightPreviewResult> = {
    request: {
        query: getDocumentNode(GET_INSIGHT_PREVIEW_GQL),
        variables: {
            input: {
                series: [
                    {
                        query: 'test query',
                        label: 'test series',
                        generatedFromCaptureGroups: true,
                        groupBy: GroupByField.AUTHOR,
                    },
                ],
                repositoryScope: { repositories: ['sourcegraph/sourcegraph'] },
                timeScope: { stepInterval: { unit: TimeIntervalStepUnit.DAY, value: 1 } },
            },
        },
    },
    result: {
        data: {
            searchInsightPreview: [
                {
                    __typename: 'SearchInsightLivePreviewSeries',
                    label: 'Foo',
                    points: [
                        { __typename: 'InsightDataPoint', diffQuery: 'type:diff', dateTime: '0', value: 100 },
                        { __typename: 'InsightDataPoint', diffQuery: 'type:diff', dateTime: '0', value: 200 },
                    ],
                },
                {
                    __typename: 'SearchInsightLivePreviewSeries',
                    label: 'Boo',
                    points: [{ __typename: 'InsightDataPoint', diffQuery: 'type:diff', dateTime: '0', value: 200 }],
                },
                {
                    __typename: 'SearchInsightLivePreviewSeries',
                    label: 'Baz',
                    points: [{ __typename: 'InsightDataPoint', diffQuery: 'type:diff', dateTime: '0', value: 500 }],
                },
                {
                    __typename: 'SearchInsightLivePreviewSeries',
                    label: 'Qux',
                    points: [{ __typename: 'InsightDataPoint', diffQuery: 'type:diff', dateTime: '0', value: 300 }],
                },
                {
                    __typename: 'SearchInsightLivePreviewSeries',
                    label: 'Corge',
                    points: [{ __typename: 'InsightDataPoint', diffQuery: 'type:diff', dateTime: '0', value: 150 }],
                },
            ],
        },
    },
}

const MOCK_SERIES: SearchBasedInsightSeries[] = [
    { id: 'series_001', name: 'test series', query: 'test query', stroke: 'var(--blue)' },
]

export const ComputeLivePreview: StoryFn = () => (
    <MockedTestProvider mocks={[mock]}>
        <div className="m-3 px-4 py-5 bg-white">
            <ComputeLivePreviewComponent
                disabled={false}
                repositories={['sourcegraph/sourcegraph']}
                series={MOCK_SERIES}
                groupBy={GroupByField.AUTHOR}
            />
        </div>
    </MockedTestProvider>
)
