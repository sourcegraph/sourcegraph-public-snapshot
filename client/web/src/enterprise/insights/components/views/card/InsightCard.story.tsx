import React from 'react'

import { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Button, Menu, MenuButton, MenuItem, MenuList } from '@sourcegraph/wildcard'

import { getLineColor, LegendItem, LegendList, ParentSize, Series } from '../../../../../charts'
import { WebStory } from '../../../../../components/WebStory'
import { SeriesChart } from '../chart'
import { SeriesBasedChartTypes } from '../types'

import * as Card from './InsightCard'

export default {
    title: 'web/insights/shared-components',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const InsightCardShowcase: Story = () => (
    <main style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem' }}>
        <section>
            <h2>Empty view</h2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Empty card" />
            </Card.Root>
        </section>

        <section>
            <h2>View with loading content</h2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Loading insight card" subtitle="View with loading content example" />
                <Card.Loading>Loading insight</Card.Loading>
            </Card.Root>
        </section>

        <section>
            <h2>View with error-like content</h2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Loading insight card" subtitle="View with errored content example" />
                <ErrorAlert error={new Error("We couldn't find code insight")} />
            </Card.Root>
        </section>

        <section>
            <h2>Card with banner content (resizing state)</h2>
            <Card.Root style={{ width: '400px', height: '400px' }}>
                <Card.Header title="Resizing insight card" subtitle="Resizing insight card" />
                <Card.Banner>Resizing</Card.Banner>
            </Card.Root>
        </section>

        <section>
            <h2>Card with insight chart</h2>
            <InsightCardWithChart />
        </section>

        <section>
            <h2>View with context action item</h2>
            <Card.Root style={{ width: 400, height: 400 }}>
                <Card.Header
                    title="Chart view and looooooong loooooooooooooooong name of insight card block"
                    subtitle="Subtitle chart description"
                >
                    <Button variant="icon" className="p-1">
                        <FilterOutlineIcon size="1rem" />
                    </Button>
                    <Menu>
                        <MenuButton variant="icon" className="p-1">
                            <DotsVerticalIcon size={16} />
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
    a: number | null
    b: number | null
    c: number | null
    x: number
}

const DATA: StandardDatum[] = [
    {
        x: 1588965700286 - 4 * 24 * 60 * 60 * 1000,
        a: 4000,
        b: 15000,
        c: 5000,
    },
    {
        x: 1588965700286 - 3 * 24 * 60 * 60 * 1000,
        a: 4000,
        b: 26000,
        c: 5000,
    },
    {
        x: 1588965700286 - 2 * 24 * 60 * 60 * 1000,
        a: 5600,
        b: 20000,
        c: 5000,
    },
    {
        x: 1588965700286 - 24 * 60 * 60 * 1000,
        a: 9800,
        b: 19000,
        c: 5000,
    },
    {
        x: 1588965700286,
        a: 6000,
        b: 17000,
        c: 5000,
    },
]

const SERIES: Series<StandardDatum>[] = [
    {
        dataKey: 'a',
        name: 'A metric',
        color: 'var(--blue)',
    },
    {
        dataKey: 'b',
        name: 'B metric',
        color: 'var(--orange)',
    },
    {
        dataKey: 'c',
        name: 'C metric',
        color: 'var(--indigo)',
    },
]

const getXValue = (datum: { x: number }) => new Date(datum.x)

function InsightCardWithChart() {
    return (
        <Card.Root style={{ width: '400px', height: '400px' }}>
            <Card.Header title="Insight with chart" subtitle="CSS migration insight chart">
                <Button variant="icon" className="p-1">
                    <FilterOutlineIcon size="1rem" />
                </Button>
                <Menu>
                    <MenuButton variant="icon" className="p-1">
                        <DotsVerticalIcon size={16} />
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
                        data={DATA}
                        series={SERIES}
                        getXValue={getXValue}
                        width={parent.width}
                        height={parent.height}
                    />
                )}
            </ParentSize>
            <LegendList className="mt-3">
                {SERIES.map(line => (
                    <LegendItem key={line.dataKey.toString()} color={getLineColor(line)} name={line.name} />
                ))}
            </LegendList>
        </Card.Root>
    )
}
