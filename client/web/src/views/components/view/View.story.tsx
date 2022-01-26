import { Menu, MenuButton, MenuItem, MenuItems, MenuPopover } from '@reach/menu-button'
import { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'
import React from 'react'
import { LineChartContent } from 'sourcegraph'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'

import * as View from '.'

export default {
    title: 'web/views',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

const standardViewProps = {
    style: { width: '400px', height: '400px' },
}

const LINE_CHART_DATA: LineChartContent<any, string> = {
    chart: 'line',
    data: [
        { x: 1588965700286 - 4 * 24 * 60 * 60 * 1000, a: null, b: null },
        { x: 1588965700286 - 3 * 24 * 60 * 60 * 1000, a: null, b: null },
        { x: 1588965700286 - 2 * 24 * 60 * 60 * 1000, a: 94, b: 200 },
        { x: 1588965700286 - 1.5 * 24 * 60 * 60 * 1000, a: 134, b: null },
        { x: 1588965700286 - 1.3 * 24 * 60 * 60 * 1000, a: null, b: 150 },
        { x: 1588965700286 - 1 * 24 * 60 * 60 * 1000, a: 134, b: 190 },
        { x: 1588965700286, a: 123, b: 170 },
    ],
    series: [
        {
            dataKey: 'a',
            name: 'A metric',
            stroke: 'var(--blue)',
            linkURLs: {
                [1588965700286 - 4 * 24 * 60 * 60 * 1000]: '#A:1st_data_point',
                [1588965700286 - 3 * 24 * 60 * 60 * 1000]: '#A:2st_data_point',
                [1588965700286 - 3 * 24 * 60 * 60 * 1000]: '#A:3rd_data_point',
                [1588965700286 - 2 * 24 * 60 * 60 * 1000]: '#A:4th_data_point',
                [1588965700286 - 1 * 24 * 60 * 60 * 1000]: '#A:5th_data_point',
            },
        },
        {
            dataKey: 'b',
            name: 'B metric',
            stroke: 'var(--warning)',
        },
    ],
    xAxis: {
        dataKey: 'x',
        scale: 'time',
        type: 'number',
    },
}

export const EmptyView: Story = () => <View.Root {...standardViewProps} title="Empty view" />

export const ViewWithChartContent: Story = () => (
    <View.Root {...standardViewProps} title="Chart view" subtitle="Subtitle chart description">
        <View.Content content={[LINE_CHART_DATA]} telemetryService={NOOP_TELEMETRY_SERVICE} />
    </View.Root>
)

export const ViewWithLoadingContent: Story = () => (
    <View.Root {...standardViewProps} title="Loading view">
        <View.LoadingContent text="Loading insight" />
    </View.Root>
)

export const ViewWithErrorLikeContent: Story = () => (
    <View.Root
        style={{ width: '400px', height: '400px' }}
        title="Error view"
        subtitle="View with errored content example"
    >
        <View.ErrorContent title="searchInsights.insight.id" error={new Error("We couldn't find code insight")} />
    </View.Root>
)

function ContextMenu() {
    return (
        <Menu>
            <MenuButton className="btn btn-icon p-1">
                <DotsVerticalIcon size={16} />
            </MenuButton>
            <MenuPopover>
                <MenuItems className="d-block position-static dropdown-menu">
                    <MenuItem onSelect={noop}>Create</MenuItem>

                    <MenuItem onSelect={noop}>Update</MenuItem>

                    <MenuItem onSelect={noop}>Delete</MenuItem>
                </MenuItems>
            </MenuPopover>
        </Menu>
    )
}

export const ViewWithContextMenu: Story = () => (
    <View.Root
        {...standardViewProps}
        title="Chart view and looooooong loooooooooooooooong name of insight card block"
        subtitle="Subtitle chart description"
        actions={
            <>
                <Button className="btn-icon p-1">
                    <FilterOutlineIcon size="1rem" />
                </Button>
                <ContextMenu />
            </>
        }
    >
        <View.Content content={[LINE_CHART_DATA]} telemetryService={NOOP_TELEMETRY_SERVICE} />
    </View.Root>
)
