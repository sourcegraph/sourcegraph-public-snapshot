import React, { useCallback, useEffect, useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button, Grid, Typography } from '..'

import { Tooltip } from './Tooltip'
import { TooltipController } from './TooltipController'

// BrandedStory already renders `<Tooltip />` so in Stories we don't render `<Tooltip />`
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
            You can <strong data-tooltip="Tooltip 1">hover me</strong> or <strong data-tooltip="Tooltip 2">me</strong>.
        </p>
    </>
)

Basic.parameters = {
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
                <Button variant="secondary" data-placement="top" data-tooltip="Tooltip on top">
                    Tooltip on top
                </Button>
            </div>
            <div>
                <Button variant="secondary" data-placement="bottom" data-tooltip="Tooltip on bottom">
                    Tooltip on bottom
                </Button>
            </div>
            <div>
                <Button variant="secondary" data-placement="left" data-tooltip="Tooltip on left">
                    Tooltip on left
                </Button>
            </div>
            <div>
                <Button variant="secondary" data-placement="right" data-tooltip="Tooltip on right">
                    Tooltip on right
                </Button>
            </div>
        </Grid>

        <Typography.H2>Max width</Typography.H2>
        <Grid columnCount={1}>
            <div>
                <Button
                    variant="secondary"
                    data-tooltip="Nulla porttitor accumsan tincidunt. Proin eget tortor risus. Quisque velit nisi, pretium ut lacinia in, elementum id enim. Donec rutrum congue leo eget malesuada."
                >
                    Tooltip with long text
                </Button>
            </div>
        </Grid>
    </>
)

Positions.parameters = {
    chromatic: {
        disable: true,
    },
}

/*
    If you take a look at the handleEvent function in useTooltipState, you can see that the listeners are being added to the 'document',
    which means any 'mouseover/click' event will cause the tooltip to disappear.
*/
export const Pinned: Story = () => {
    const clickElement = useCallback((element: HTMLElement | null) => {
        if (element) {
            // The tooltip takes some time to set-up.
            // hence we need to delay the click by some ms.
            setTimeout(() => {
                element.click()
            }, 10)
        }
    }, [])

    return (
        <>
            <span data-tooltip="My tooltip" ref={clickElement}>
                Example
            </span>
            <p>
                <small>
                    (A pinned tooltip is shown when the target element is rendered, without any user interaction
                    needed.)
                </small>
            </p>
        </>
    )
}

Pinned.parameters = {
    chromatic: {
        // Chromatic pauses CSS animations by default and resets them to their initial state
        pauseAnimationAtEnd: true,
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

const ForceUpdateTooltip = () => {
    const [copied, setCopied] = useState<boolean>(false)

    useEffect(() => {
        TooltipController.forceUpdate()
    }, [copied])

    const onClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        event.preventDefault()

        setCopied(true)

        setTimeout(() => {
            setCopied(false)
        }, 1500)
    }

    return (
        <>
            <Typography.H2>
                Force update tooltip with <code>TooltipController.forceUpdate()</code>
            </Typography.H2>
            <p>
                <Button variant="primary" onClick={onClick} data-tooltip={copied ? 'Copied!' : 'Click to copy'}>
                    Button
                </Button>
            </p>
        </>
    )
}

export const Controller: Story = () => <ForceUpdateTooltip />

Controller.parameters = {
    chromatic: {
        disable: true,
    },
}
