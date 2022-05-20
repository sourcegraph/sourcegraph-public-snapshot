import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button, Typography } from '..'

import { Tooltip } from './Tooltip'

// BrandedStory already renders `<Tooltip />` so in Stories we don't render `<Tooltip />`
const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="p-5">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/TooltipNew',

    decorators: [decorator],

    parameters: {
        component: Tooltip,
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=3131%3A38534',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=3131%3A38727',
            },
        ],
    },
}

export default config

export const Examples: Story = () => (
    <>
        <Typography.H1>Using Popover</Typography.H1>
        <div>
            <Tooltip title="First tooltip">
                <Button variant="secondary">First tooltip</Button>
            </Tooltip>
            <Tooltip title="Other toolip">
                <Button variant="secondary">Other tooltip</Button>
            </Tooltip>
        </div>
    </>
)
