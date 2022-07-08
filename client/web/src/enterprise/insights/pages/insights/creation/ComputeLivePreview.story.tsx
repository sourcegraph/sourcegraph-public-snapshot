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
                            value: 100,
                            dateTime: new Date('2020-01-01'),
                            link,
                        },
                    ],
                    getLinkURL: datum => datum.link,
                    getYValue: datum => datum.value,
                    getXValue: datum => datum.dateTime,
                },
                {
                    id: 'bar',
                    name: 'Bar',
                    color: 'var(--blue)',
                    data: [
                        {
                            value: 200,
                            dateTime: new Date('2020-02-01'),
                            link,
                        },
                    ],
                    getLinkURL: datum => datum.link,
                    getYValue: datum => datum.value,
                    getXValue: datum => datum.dateTime,
                },
                {
                    id: 'baz',
                    name: 'Baz',
                    color: 'var(--green)',
                    data: [
                        {
                            value: 150,
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
        <div className="m-3 px-4 py-5 bg-white">
            <ComputeLivePreviewComponent disabled={false} repositories="sourcegraph/sourcegraph" series={[]} />
        </div>
    </CodeInsightsBackendStoryMock>
)
