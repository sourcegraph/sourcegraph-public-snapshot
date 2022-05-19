import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Tooltip } from '.'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="p-5">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Tooltip',

    decorators: [decorator],

    parameters: {
        component: 'Tooltip',
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

export const Basic: Story = () => (
    <>
        <p>
            You can{' '}
            <Tooltip.Root>
                <Tooltip.Trigger>
                    <strong>hover me</strong>
                </Tooltip.Trigger>
                <Tooltip.Content>Tooltip 1</Tooltip.Content>
            </Tooltip.Root>
            or{' '}
            <Tooltip.Root>
                <Tooltip.Trigger>
                    <strong>me</strong>
                </Tooltip.Trigger>
                <Tooltip.Content>Tooltip 2</Tooltip.Content>
            </Tooltip.Root>
            .
        </p>
    </>
)

Basic.parameters = {
    chromatic: {
        disable: true,
    },
}
