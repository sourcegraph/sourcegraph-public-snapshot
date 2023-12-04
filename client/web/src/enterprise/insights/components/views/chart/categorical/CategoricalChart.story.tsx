import type { Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../../components/WebStory'

import { CategoricalBasedChartTypes, CategoricalChart } from './CategoricalChart'

const StoryConfig: Meta = {
    title: 'web/insights/views/CategoricalChart',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default StoryConfig

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    linkURL: string
}

const LANGUAGE_USAGE_DATA: LanguageUsageDatum[] = [
    {
        name: 'JavaScript',
        value: 422,
        fill: '#f1e05a',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        name: 'CSS',
        value: 273,
        fill: '#563d7c',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        name: 'HTML',
        value: 129,
        fill: '#e34c26',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        name: 'Markdown',
        value: 35,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]

const getValue = (datum: LanguageUsageDatum) => datum.value
const getColor = (datum: LanguageUsageDatum) => datum.fill
const getLink = (datum: LanguageUsageDatum) => datum.linkURL
const getName = (datum: LanguageUsageDatum) => datum.name

export const CategoricalPieChart: StoryFn = () => (
    <CategoricalChart
        type={CategoricalBasedChartTypes.Pie}
        width={400}
        height={400}
        data={LANGUAGE_USAGE_DATA}
        getDatumName={getName}
        getDatumValue={getValue}
        getDatumColor={getColor}
        getDatumLink={getLink}
    />
)
