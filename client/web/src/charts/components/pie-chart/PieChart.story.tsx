import React from 'react'

import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { PieChart } from './PieChart'

export default {
    title: 'web/charts/pie',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

interface LanguageUsageDatum {
    name: string,
    value: number,
    fill: string,
    linkURL: string,
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

export const PieChartVitrina: Story = () => (
    <PieChart<LanguageUsageDatum>
        width={400}
        height={400}
        data={LANGUAGE_USAGE_DATA}
        getDatumName={getName}
        getDatumValue={getValue}
        getDatumColor={getColor}
        getDatumLink={getLink}
    />
)
