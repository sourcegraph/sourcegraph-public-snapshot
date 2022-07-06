import { Meta, Story } from '@storybook/react'

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

const codeInsightsBackend = {
    getInsightPreviewContent: (): Promise<SeriesChartContent<BackendInsightDatum>> =>
        Promise.resolve({
            series: [
                {
                    id: 'foo',
                    name: 'Foo',
                    color: 'var(--red)',
                    data: [
                        {
                            value: 10,
                            dateTime: new Date('2020-01-01'),
                            link,
                        },
                        {
                            value: 20,
                            dateTime: new Date('2020-02-01'),
                            link,
                        },
                    ],
                    getLinkURL: datum => datum.link,
                    getYValue: datum => datum.value,
                    getXValue: datum => datum.dateTime,
                },
                {
                    id: 'bar',
                    name: 'Boo',
                    color: 'var(--blue)',
                    data: [
                        {
                            value: 20,
                            dateTime: new Date('2020-02-01'),
                            link,
                        },
                    ],
                    getLinkURL: datum => datum.link,
                    getYValue: datum => datum.value,
                    getXValue: datum => datum.dateTime,
                },
            ],
        }),
}

export const ComputeLivePreview: Story = () => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
        <ComputeLivePreviewComponent
            disabled={false}
            repositories="sourcegraph/sourcegraph"
            stepValue="2"
            step="weeks"
            series={[]}
        />
    </CodeInsightsBackendStoryMock>
)
