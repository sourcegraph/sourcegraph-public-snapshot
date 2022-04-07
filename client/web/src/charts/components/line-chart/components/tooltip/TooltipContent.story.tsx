import React from 'react'

import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { Series } from '../../../../types'

import { MinimumPointInfo, TooltipContent } from './TooltipContent'

const StoryConfig: Meta = {
    title: 'web/charts/tooltip',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default StoryConfig

interface Datum {
    a: number
    b: number
    c: number
    x: number
}

const SERIES: Series<Datum>[] = [
    {
        dataKey: 'a',
        name: 'A metric',
        color: 'var(--blue)',
    },
    {
        dataKey: 'c',
        name: 'C metric',
        color: 'var(--purple)',
    },
    {
        dataKey: 'b',
        name: 'B metric',
        color: 'var(--warning)',
    },
]

const LONG_NAMED_SERIES: Series<Datum>[] = [
    {
        dataKey: 'a',
        name: 'In_days,_when_all_earthly+impressions_Where_utter_novelty_for_me',
        color: 'var(--blue)',
    },
    {
        dataKey: 'c',
        name: 'And_looks_of_maids_and_noise_of_groves,_And_nightingale’s_plea',
        color: 'var(--purple)',
    },
    {
        dataKey: 'b',
        name: 'When_highly_elevated_senses,_The_love,_the_liberty,_the_pride',
        color: 'var(--warning)',
    },
]

const ACTIVE_POINT: MinimumPointInfo<Datum> = {
    seriesKey: 'a',
    value: 200,
    time: new Date('2020-05-07T19:21:40.286Z'),
    datum: {
        x: 1588879300286,
        a: 134,
        b: 190,
        c: 190,
    },
}

export const TooltipLayouts: Story = () => (
    <div className="d-flex flex-column" style={{ gap: 20 }}>
        <div>
            <h2>Regular tooltip</h2>
            <TooltipContent stacked={false} series={SERIES} activePoint={ACTIVE_POINT} />
        </div>

        <div>
            <h2>With stacked value</h2>
            <TooltipContent stacked={true} series={SERIES} activePoint={ACTIVE_POINT} />
        </div>

        <div>
            <h2>With long named series</h2>
            <TooltipContent stacked={true} series={LONG_NAMED_SERIES} activePoint={ACTIVE_POINT} />
        </div>
    </div>
)
