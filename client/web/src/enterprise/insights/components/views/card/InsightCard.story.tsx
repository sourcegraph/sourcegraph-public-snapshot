import { mdiFilterOutline, mdiDotsVertical } from '@mdi/js'
import type { Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import {
    Button,
    Menu,
    MenuButton,
    MenuItem,
    MenuList,
    H2,
    Icon,
    LegendItem,
    LegendList,
    ParentSize,
    type Series,
    ErrorAlert,
} from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../components/WebStory'
import { useSeriesToggle } from '../../../../../insights/utils/use-series-toggle'
import { SeriesBasedChartTypes, SeriesChart } from '../chart'

import * as Card from './InsightCard'

const meta: Meta = {
    title: 'web/insights/shared-components',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default meta

export const InsightCardShowcase: StoryFn = () => (
    <main style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem' }}>
        <section>
            <H2>Empty view</H2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Empty card" />
            </Card.Root>
        </section>

        <section>
            <H2>View with loading content</H2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Loading insight card" subtitle="View with loading content example" />
                <Card.Loading>Loading insight</Card.Loading>
            </Card.Root>
        </section>

        <section>
            <H2>View with error-like content</H2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Loading insight card" subtitle="View with errored content example" />
                <ErrorAlert error={new Error("We couldn't find code insight")} />
            </Card.Root>
        </section>

        <section>
            <H2>Card with banner content (resizing state)</H2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Resizing insight card" subtitle="Resizing insight card" />
                <Card.Banner>Resizing</Card.Banner>
            </Card.Root>
        </section>

        <section>
            <H2>Card with insight chart</H2>
            <InsightCardWithChart />
        </section>

        <section>
            <H2>View with context action item</H2>
            <Card.Root style={{ width: 400, height: 400 }}>
                <Card.Header
                    title="Chart view and looooooong loooooooooooooooong name of insight card block"
                    subtitle="Subtitle chart description"
                >
                    <Button variant="icon" className="p-1">
                        <Icon
                            svgPath={mdiFilterOutline}
                            inline={false}
                            aria-label="Filters"
                            height="1rem"
                            width="1rem"
                        />
                    </Button>
                    <Menu>
                        <MenuButton variant="icon" className="p-1">
                            <Icon
                                svgPath={mdiDotsVertical}
                                inline={false}
                                aria-label="Options"
                                height={16}
                                width={16}
                            />
                        </MenuButton>
                        <MenuList>
                            <MenuItem onSelect={noop}>Create</MenuItem>
                            <MenuItem onSelect={noop}>Update</MenuItem>
                            <MenuItem onSelect={noop}>Delete</MenuItem>
                        </MenuList>
                    </Menu>
                </Card.Header>
            </Card.Root>
        </section>
    </main>
)

interface StandardDatum {
    value: number
    x: number
}

const getXValue = (datum: StandardDatum): Date => new Date(datum.x)
const getYValue = (datum: StandardDatum): number => datum.value

const SERIES: Series<StandardDatum>[] = [
    {
        id: 'series_001',
        data: [
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 4000 },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 4000 },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5600 },
            { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 9800 },
            { x: 1588965700286, value: 6000 },
        ],
        name: 'A metric',
        color: 'var(--blue)',
        getXValue,
        getYValue,
    },
    {
        id: 'series_002',
        data: [
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 15000 },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 26000 },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 20000 },
            { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 19000 },
            { x: 1588965700286, value: 17000 },
        ],
        name: 'B metric',
        color: 'var(--orange)',
        getXValue,
        getYValue,
    },
    {
        id: 'series_003',
        data: [
            { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, value: 5000 },
            { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, value: 5000 },
            { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, value: 5000 },
            { x: 1588965700286 - 24 * 60 * 60 * 1000, value: 5000 },
            { x: 1588965700286, value: 5000 },
        ],
        name: 'C metric',
        color: 'var(--indigo)',
        getXValue,
        getYValue,
    },
]

function InsightCardWithChart() {
    const seriesToggleState = useSeriesToggle()

    return (
        <Card.Root style={{ width: '400px', height: '400px' }}>
            <Card.Header title="Insight with chart" subtitle="CSS migration insight chart">
                <Button variant="icon" className="p-1">
                    <Icon svgPath={mdiFilterOutline} inline={false} aria-label="Options" height="1rem" width="1rem" />
                </Button>
                <Menu>
                    <MenuButton variant="icon" className="p-1">
                        <Icon svgPath={mdiDotsVertical} inline={false} aria-label="Filters" height={16} width={16} />
                    </MenuButton>
                    <MenuList>
                        <MenuItem onSelect={noop}>Create</MenuItem>
                        <MenuItem onSelect={noop}>Update</MenuItem>
                        <MenuItem onSelect={noop}>Delete</MenuItem>
                    </MenuList>
                </Menu>
            </Card.Header>
            <ParentSize>
                {parent => (
                    <SeriesChart
                        type={SeriesBasedChartTypes.Line}
                        series={SERIES}
                        width={parent.width}
                        height={parent.height}
                        seriesToggleState={seriesToggleState}
                    />
                )}
            </ParentSize>
            <LegendList className="mt-3">
                {SERIES.map(line => (
                    <LegendItem key={line.id} color={line.color} name={line.name} />
                ))}
            </LegendList>
        </Card.Root>
    )
}
