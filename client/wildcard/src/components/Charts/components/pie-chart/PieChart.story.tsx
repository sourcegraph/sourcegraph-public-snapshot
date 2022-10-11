import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { H2, Text } from '../../../Typography'

import { PieChart } from './PieChart'

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
}

const getValue = (datum: LanguageUsageDatum) => datum.value
const getColor = (datum: LanguageUsageDatum) => datum.fill
const getLink = (datum: LanguageUsageDatum) => datum.linkURL
const getName = (datum: LanguageUsageDatum) => datum.name

export const PieChartDemo: Story = () => (
    <main
        style={{
            display: 'flex',
            flexWrap: 'wrap',
            rowGap: 40,
            columnGap: 40,
            paddingBottom: 40,
        }}
    >
        <PlainPieChartExample />
        <ManyGroupsPieChartExample />
    </main>
)

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

const PlainPieChartExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Plain pie chart</H2>

        <Text>
            Standard PieChart example. All pie chart parts (arcs) are focusable if links for the pie chart are provided
            through the getDatumLink prop.
        </Text>

        <PieChart<LanguageUsageDatum>
            width={400}
            height={400}
            data={LANGUAGE_USAGE_DATA}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
        />
    </section>
)

const MANY_LANGUAGES_DATA: LanguageUsageDatum[] = [
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
        name: 'Julia',
        value: 40,
        fill: '#268ee3',
        linkURL: 'https://en.wikipedia.org/wiki/Julia',
    },
    {
        name: 'Rust',
        value: 35,
        fill: '#e37b26',
        linkURL: 'https://en.wikipedia.org/wiki/rust',
    },
    {
        name: 'C#',
        value: 32,
        fill: '#ad26e3',
        linkURL: 'https://en.wikipedia.org/wiki/c#',
    },
    {
        name: 'C++',
        value: 30,
        fill: '#e32626',
        linkURL: 'https://en.wikipedia.org/wiki/c++',
    },
    {
        name: 'Markdown',
        value: 20,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]

const ManyGroupsPieChartExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Many arcs example</H2>

        <Text>
            The pie chart supports bringing in the front hovered/focused arc element annotation tooltip. In the case of
            many arcs, it might be useful, but we suggest keeping the number of arcs small and group small value groups
            in one "Other" group.
        </Text>

        <PieChart<LanguageUsageDatum>
            width={400}
            height={400}
            data={MANY_LANGUAGES_DATA}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
        />
    </section>
)
