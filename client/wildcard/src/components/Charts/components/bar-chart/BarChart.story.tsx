import { Meta, Story } from '@storybook/react'
import { ParentSize } from '@visx/responsive'
import { ResizableBox } from 'react-resizable'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Badge } from '../../../Badge'
import { H2, Text } from '../../../Typography'

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
        value: 129,
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
        value: 422,
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
const getGroup = (datum: LanguageUsageDatum) => datum.group

export const BarChartDemo: Story = () => (
    <main
        style={{
            display: 'flex',
            flexWrap: 'wrap',
            rowGap: 40,
            columnGap: 20,
            paddingBottom: 40,
        }}
    >
        <PlainBarChartExample />
        <SortedBarChartExample />
        <GroupedBarExample />
        <StackedBarExample />
        <ManyBarsExample />
        <CustomDimmedColor />
    </main>
)

const PlainBarChartExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Plain bar chart</H2>

        <Text>
            By default bar chart uses a bar name as a group name, so in this example, each bar has its own group (each
            bar is independent). See grouped bar example for bars grouping.
        </Text>

        <BarChart
            width={400}
            height={400}
            data={LANGUAGE_USAGE_DATA}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
            getDatumHover={datum => `custom text for ${datum.name}`}
        />
    </section>
)

const SortedBarChartExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Sorted bar chart</H2>

        <Text>This is the default bar chart sorted by descending value.</Text>

        <BarChart
            width={400}
            height={400}
            data={LANGUAGE_USAGE_DATA}
            sortByValue={true}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
            getDatumHover={datum => `custom text for ${datum.name}`}
        />
    </section>
)

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
        name: 'Markdown',
        value: 300,
        fill: '#083fa1',
        linkURL: 'https://en.wikipedia.org/wiki/Markdown',
    },
]

const GroupedBarExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Grouped bar chart</H2>

        <Text>It's possible to group (categories) bars by group name. You can do it with the `getCategory` prop.</Text>

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
    </section>
)

const StackedBarExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Stacked bar chart</H2>

        <Text>
            <Badge variant="merged">Experimental</Badge> You can stack bars which are placed in one category (group).
        </Text>

        <BarChart
            width={400}
            height={400}
            data={LANGUAGE_USAGE_DATA}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumLink={getLink}
            getDatumHover={datum => `custom text for ${datum.name}`}
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

const ManyBarsExample = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Smart labels UI</H2>

        <Text style={{ maxWidth: 400, minWidth: 400 }}>
            <Badge variant="merged">Experimental</Badge> Try to resize charts (drag bottom right corner), note that
            labels rotate if they don't have enough space.
        </Text>

        <ResizableBox width={400} height={400} axis="both" minConstraints={[200, 200]} className="p-3">
            <ParentSize debounceTime={0}>
                {parent => (
                    <BarChart
                        width={parent.width}
                        height={parent.height}
                        data={MANY_LANGUAGES_DATA}
                        getDatumName={getName}
                        getDatumValue={getValue}
                        getDatumColor={getColor}
                        getDatumLink={getLink}
                        getDatumHover={datum => `custom text for ${datum.name}`}
                    />
                )}
            </ParentSize>
        </ResizableBox>
    </section>
)

const CustomDimmedColor = () => (
    <section style={{ flexBasis: 0 }}>
        <H2>Dimmed colors</H2>

        <Text style={{ maxWidth: 400, minWidth: 400 }}>
            You can specify any dimmed colors for the non-active bars. (see bar chart README.md for more details about
            chart colours.
        </Text>

        <BarChart
            width={400}
            height={400}
            data={LANGUAGE_USAGE_DATA}
            getDatumName={getName}
            getDatumValue={getValue}
            getDatumColor={getColor}
            getDatumFadeColor={() => 'var(--blue)'}
            getDatumLink={getLink}
        />
    </section>
)
