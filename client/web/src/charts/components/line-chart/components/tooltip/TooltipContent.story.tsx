import { Meta, Story } from '@storybook/react'

import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../components/WebStory'
import { Series } from '../../../../types'

import { MinimumPointInfo, TooltipContent } from './TooltipContent'

const StoryConfig: Meta = {
    title: 'web/charts/tooltip',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
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

const ACTIVE_POINT: MinimumPointInfo<Datum> = {
    seriesId: 'a',
    value: 200,
    time: new Date('2020-05-07T19:21:40.286Z'),
}

export const TooltipLayouts: Story = () => (
    <div className="d-flex flex-column" style={{ gap: 20 }}>
        <div>
            <Typography.H2>Regular tooltip</Typography.H2>
            <TooltipContent stacked={false} series={SERIES} activePoint={ACTIVE_POINT} />
        </div>

        <div>
            <Typography.H2>With stacked value</Typography.H2>
            <TooltipContent stacked={true} series={SERIES} activePoint={ACTIVE_POINT} />
        </div>

        <div>
            <Typography.H2>With long named series</Typography.H2>
            <TooltipContent stacked={true} series={LONG_NAMED_SERIES} activePoint={ACTIVE_POINT} />
        </div>
    </div>
)
