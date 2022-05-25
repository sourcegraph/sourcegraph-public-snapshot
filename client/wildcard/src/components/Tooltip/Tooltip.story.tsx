import React, { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../Button'
import { Grid } from '../Grid'
import { Typography } from '../Typography'

import { Tooltip } from '.'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="p-5">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Tooltip',

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

export const Basic: Story = () => (
    <>
        <p>
            You can{' '}
            <Tooltip content="Tooltip 1">
                <strong>hover me</strong>
            </Tooltip>{' '}
            or{' '}
            <Tooltip content="Tooltip 2">
                <strong>me</strong>
            </Tooltip>
            .
        </p>
    </>
)

Basic.parameters = {
    chromatic: {
        disable: true,
    },
}

export const DisabledTrigger: Story = () => (
    <Tooltip content="Tooltip still works properly">
        <Button variant="primary" disabled={true} style={{ pointerEvents: 'none' }}>
            Disabled Button 🚫
        </Button>
    </Tooltip>
)

DisabledTrigger.parameters = {
    chromatic: {
        disable: true,
    },
}

export const Pinned: Story = () => (
    <>
        <Tooltip content="My tooltip" defaultOpen={true}>
            Example
        </Tooltip>

        <p>
            <small>
                (A pinned tooltip is shown when the target element is rendered, without any user interaction needed.)
            </small>
        </p>
    </>
)

Pinned.parameters = {
    chromatic: {
        disable: true,
    },
}

export const Positions: Story = () => (
    <>
        <Typography.H1>Tooltip</Typography.H1>
        <Typography.H2>Positions</Typography.H2>

        <Grid columnCount={4}>
            <div>
                <Tooltip content="Tooltip on top" placement="top">
                    <Button variant="secondary">Top</Button>
                </Tooltip>
            </div>

            <div>
                <Tooltip content="Tooltip on bottom" placement="bottom">
                    <Button variant="secondary">Bottom</Button>
                </Tooltip>
            </div>

            <div>
                <Tooltip content="Tooltip on left" placement="left">
                    <Button variant="secondary">Left</Button>
                </Tooltip>
            </div>

            <div>
                <Tooltip content="Tooltip on right" placement="right">
                    <Button variant="secondary">Right</Button>
                </Tooltip>
            </div>
        </Grid>

        <Typography.H2>Max width</Typography.H2>
        <Grid columnCount={1}>
            <div>
                <Tooltip content="Nulla porttitor accumsan tincidunt. Proin eget tortor risus. Quisque velit nisi, pretium ut lacinia in, elementum id enim. Donec rutrum congue leo eget malesuada.">
                    <Button variant="secondary">Tooltip with long text</Button>
                </Tooltip>
            </div>
        </Grid>
    </>
)

Positions.parameters = {
    chromatic: {
        disable: true,
    },
}

export const UpdateText: Story = () => {
    const [copied, setCopied] = useState<boolean>(false)

    function onClick(event: React.MouseEvent<HTMLButtonElement>) {
        setCopied(true)
        setTimeout(() => setCopied(false), 1500)
        event.target.dispatchEvent(new Event('focus'))
    }

    return (
        <Tooltip content={copied ? 'Copied!' : 'Click to copy.'}>
            <Button variant="primary" onClick={onClick}>
                Copy
            </Button>
        </Tooltip>
    )
}

UpdateText.parameters = {
    chromatic: {
        disable: true,
    },
}
