import { Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import DotsVerticalIcon from 'mdi-react/DotsVerticalIcon'
import FilterOutlineIcon from 'mdi-react/FilterOutlineIcon'

import { Button, Menu, MenuButton, MenuItem, MenuList, Position } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'
import { LINE_CHART_CONTENT_MOCK } from '../../mocks/charts-content'
import * as View from '../view'

import { ViewGrid } from './ViewGrid'

export default {
    title: 'web/views',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const SimpleViewGrid: Story = () => (
    <ViewGrid viewIds={['1', '2', '3']}>
        <View.Root key="1" title="Empty view" />

        <View.Root key="2" title="View with chart">
            <View.Content content={[LINE_CHART_CONTENT_MOCK]} />
        </View.Root>

        <View.Root
            key="3"
            title="Chart view"
            subtitle="Subtitle chart description"
            actions={
                <>
                    <Button variant="icon" className="p-1">
                        <FilterOutlineIcon size="1rem" />
                    </Button>
                    <ContextMenu />
                </>
            }
        >
            <View.Content content={[LINE_CHART_CONTENT_MOCK]} />
        </View.Root>
    </ViewGrid>
)

function ContextMenu() {
    return (
        <Menu>
            <MenuButton variant="icon" className="p-1">
                <DotsVerticalIcon size={16} />
            </MenuButton>
            <MenuList position={Position.bottomEnd}>
                <MenuItem onSelect={noop}>Create</MenuItem>

                <MenuItem onSelect={noop}>Update</MenuItem>

                <MenuItem onSelect={noop}>Delete</MenuItem>
            </MenuList>
        </Menu>
    )
}
