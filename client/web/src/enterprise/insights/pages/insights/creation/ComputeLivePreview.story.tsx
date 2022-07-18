import { Meta, Story } from '@storybook/react'

import { GroupByField } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../CodeInsightsBackendStoryMock'
import { BackendInsightDatum, SeriesChartContent } from '../../../core'

import { ComputeLivePreview as ComputeLivePreviewComponent } from './ComputeLivePreview'

const defaultStory: Meta = {
    title: 'web/insights/creation-ui/ComputeLivePreview',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory

const link = 'https://sourcegraph.com'

const mockSeriesData = [
    {
        id: 'Foo',
        color: 'yellow',
        value: 241,
    },
    {
        id: 'Boo',
        color: 'grape',
        value: 148,
    },
    {
        id: 'Baz',
        color: 'cyan',
        value: 87,
    },
    {
        id: 'Qux',
        color: 'yellow',
        value: 168,
    },
    {
        id: 'Quux',
        color: 'grape',
        value: 130,
    },
    {
        id: 'Corge',
        color: 'cyan',
        value: 118,
    },
].map(series => ({
    id: series.id,
    name: series.id,
    color: `var(--oc-${series.color}-9)`,
    data: [
        {
            value: series.value,
            dateTime: new Date('2020-01-01'),
            link,
        },
    ],
    getLinkURL: (datum: any) => datum.link,
    getYValue: (datum: any) => datum.value,
    getXValue: (datum: any) => datum.dateTime,
    getCategory: (datum: any) => datum.category,
}))

const codeInsightsBackend = {
    getInsightPreviewContent: (): Promise<SeriesChartContent<BackendInsightDatum>> =>
        Promise.resolve({
            series: mockSeriesData,
        }),
}

export const ComputeLivePreview: Story = () => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
        <div className="m-3 px-4 py-5 bg-white">
            <ComputeLivePreviewComponent
                disabled={false}
                repositories="sourcegraph/sourcegraph"
                series={[]}
                groupBy={GroupByField.AUTHOR}
            />
        </div>
    </CodeInsightsBackendStoryMock>
)
