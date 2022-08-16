import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { BarChart } from './BarChart'

const StoryConfig: Meta = {
    title: 'wildcard/Charts',
    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        chromatic: { disableSnapshots: false, enableDarkMode: true },
    },
}

export default StoryConfig

interface LanguageUsageDatum {
    name: string
    value: number
    fill: string
    linkURL: string
    group?: string
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

const LANGUAGE_USAGE_GROUPED_BY_REPO_DATA: LanguageUsageDatum[] = [
    {
        group: 'Sourcegraph',
        name: 'JavaScript',
        value: 422,
        fill: '#f1e05a',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        group: 'Sourcegraph',
        name: 'CSS',
        value: 273,
        fill: '#563d7c',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        group: 'Sourcegraph',
        name: 'HTML',
        value: 20,
        fill: '#e34c26',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        group: 'Sourcegraph',
        name: 'Markdown',
        value: 135,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
    {
        group: 'About',
        name: 'JavaScript',
        value: 300,
        fill: '#f1e05a',
        linkURL: 'https://en.wikipedia.org/wiki/JavaScript',
    },
    {
        group: 'About',
        name: 'CSS',
        value: 150,
        fill: '#563d7c',
        linkURL: 'https://en.wikipedia.org/wiki/CSS',
    },
    {
        group: 'About',
        name: 'HTML',
        value: 390,
        fill: '#e34c26',
        linkURL: 'https://en.wikipedia.org/wiki/HTML',
    },
    {
        // group: 'About',
        name: 'Markdown',
        value: 300,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]

const getValue = (datum: LanguageUsageDatum) => datum.value
const getColor = (datum: LanguageUsageDatum) => datum.fill
const getLink = (datum: LanguageUsageDatum) => datum.linkURL
const getName = (datum: LanguageUsageDatum) => datum.name
const getGroup = (datum: LanguageUsageDatum) => datum.group

export const BarChartVitrina: Story = () => (
    <div className="d-flex flex-wrap" style={{ gap: 20 }}>
        <BarChart
            width={400}
            height={400}
            data={LANGUAGE_USAGE_DATA}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
        />
        <BarChart
            width={400}
            height={400}
            data={LANGUAGE_USAGE_GROUPED_BY_REPO_DATA}
            getCategory={getGroup}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
        />
        <BarChart
            stacked={true}
            width={400}
            height={400}
            data={LANGUAGE_USAGE_GROUPED_BY_REPO_DATA}
            getCategory={getGroup}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
        />
    </div>
)
