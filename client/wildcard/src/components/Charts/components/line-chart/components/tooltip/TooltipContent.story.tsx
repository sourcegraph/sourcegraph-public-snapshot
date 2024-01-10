import type { Meta, StoryFn } from '@storybook/react'

import { BrandedStory } from '../../../../../../stories/BrandedStory'
import { H2 } from '../../../../../Typography'
import type { Series } from '../../../../types'
import { getSeriesData, type SeriesWithData } from '../../utils'

import { type MinimumPointInfo, TooltipContent } from './TooltipContent'

const StoryConfig: Meta = {
    title: 'wildcard/Charts/Core',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
}

export default StoryConfig

interface Datum {
    value: number
    x: number
}

const getXValue = (datum: Datum): Date => new Date(datum.x)
const getYValue = (datum: Datum): number => datum.value

const SERIES: Series<Datum>[] = [
    {
        id: 'a',
        data: [{ x: 1588879300286, value: 134 }],
        name: 'A metric',
        color: 'var(--blue)',
        getXValue,
        getYValue,
    },
    {
        id: 'c',
        data: [{ x: 1588879300286, value: 190 }],
        name: 'C metric',
        color: 'var(--purple)',
        getXValue,
        getYValue,
    },
    {
        id: 'b',
        data: [{ x: 1588879300286, value: 190 }],
        name: 'B metric',
        color: 'var(--warning)',
        getXValue,
        getYValue,
    },
]

const SERIES_WITH_DATA: SeriesWithData<Datum>[] = getSeriesData({ series: SERIES, stacked: false })

const LONG_NAMED_SERIES: Series<Datum>[] = [
    {
        id: 'a',
        data: [{ x: 1588879300286, value: 134 }],
        name: 'In_days,_when_all_earthly+impressions_Where_utter_novelty_for_me',
        color: 'var(--blue)',
        getXValue,
        getYValue,
    },
    {
        id: 'c',
        data: [{ x: 1588879300286, value: 190 }],
        name: 'And_looks_of_maids_and_noise_of_groves,_And_nightingale’s_plea',
        color: 'var(--purple)',
        getXValue,
        getYValue,
    },
    {
        id: 'b',
        data: [{ x: 1588879300286, value: 190 }],
        name: 'When_highly_elevated_senses,_The_love,_the_liberty,_the_pride',
        color: 'var(--warning)',
        getXValue,
        getYValue,
    },
]

const LONG_NAMED_SERIES_WITH_DATA: SeriesWithData<Datum>[] = getSeriesData({
    series: LONG_NAMED_SERIES,
    stacked: false,
})

const ACTIVE_POINT: MinimumPointInfo = {
    seriesId: 'a',
    yValue: 200,
    xValue: new Date('2020-05-07T19:21:40.286Z'),
}

export const TooltipLayoutDemo: StoryFn = () => (
    <div className="d-flex flex-column" style={{ gap: 20 }}>
        <div>
            <H2>Regular tooltip</H2>
            <TooltipContent stacked={false} series={SERIES_WITH_DATA} activePoint={ACTIVE_POINT} />
        </div>

        <div>
            <H2>With stacked value</H2>
            <TooltipContent stacked={true} series={SERIES_WITH_DATA} activePoint={ACTIVE_POINT} />
        </div>

        <div>
            <H2>With long named series</H2>
            <TooltipContent stacked={true} series={LONG_NAMED_SERIES_WITH_DATA} activePoint={ACTIVE_POINT} />
        </div>
    </div>
)
