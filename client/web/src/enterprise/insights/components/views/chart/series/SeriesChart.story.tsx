import type { Meta, StoryFn } from '@storybook/react'

import type { Series } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../../components/WebStory'
import { useSeriesToggle } from '../../../../../../insights/utils/use-series-toggle'

import { SeriesBasedChartTypes, SeriesChart } from './SeriesChart'

const StoryConfig: Meta = {
    title: 'web/insights/views/SeriesChart',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default StoryConfig

interface StandardDatum {
    value: number
    link: string
    x: number
}

const getXValue = (datum: StandardDatum): Date => new Date(datum.x)
const getYValue = (datum: StandardDatum): number => datum.value
const getLinkURL = (datum: StandardDatum): string => datum.link

const SERIES: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000, link: 'https://google.com/search' },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 4000, link: 'https://google.com/search' },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5600, link: 'https://google.com/search' },
            { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 9800, link: 'https://google.com/search' },
            { x: 1588965700286, value: 6000, link: 'https://google.com/search' },
        ],
        name: 'A metric',
        color: 'var(--blue)',
        getLinkURL,
        getXValue,
        getYValue,
    },
    {
        id: 'series_002',
        data: [
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 15000, link: 'https://yandex.com/search' },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 26000, link: 'https://yandex.com/search' },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 20000, link: 'https://yandex.com/search' },
            { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 19000, link: 'https://yandex.com/search' },
            { x: 1588965700286, value: 17000, link: 'https://yandex.com/search' },
        ],
        name: 'B metric',
        color: 'var(--warning)',
        getLinkURL,
        getXValue,
        getYValue,
    },
    {
        id: 'series_003',
        data: [
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 5000, link: 'https://twitter.com/search' },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 5000, link: 'https://twitter.com/search' },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5000, link: 'https://twitter.com/search' },
            { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, value: 5000, link: 'https://twitter.com/search' },
            { x: 1588965700286, value: 5000, link: 'https://twitter.com/search' },
        ],
        name: 'C metric',
        color: 'var(--green)',
        getLinkURL,
        getXValue,
        getYValue,
    },
]

export const SeriesLineChart: StoryFn = () => {
    const seriesToggleState = useSeriesToggle()

    return (
        <SeriesChart
            type={SeriesBasedChartTypes.Line}
            series={SERIES}
            stacked={false}
            width={400}
            height={400}
            seriesToggleState={seriesToggleState}
        />
    )
}
