import { useState } from 'react'

import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { Button, Grid, Code, Text, Input } from '..'
import { BrandedStory } from '../../stories/BrandedStory'

import { Tooltip } from '.'

const decorator: Decorator = story => <BrandedStory>{() => <div className="p-5">{story()}</div>}</BrandedStory>

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

export const Basic: StoryFn = () => (
    <Text>
        You can{' '}
        <Tooltip content="Tooltip 1">
            <strong>hover me</strong>
        </Tooltip>{' '}
        or{' '}
        <Tooltip content="Tooltip 2">
            <strong>me</strong>
        </Tooltip>
        .
    </Text>
)

export const Conditional: StoryFn = () => {
    const [clicked, setClicked] = useState<boolean>(false)

    function onClick() {
        setClicked(true)
        setTimeout(() => setClicked(false), 1500)
    }

    return (
        <Grid columnCount={1}>
            <div>
                <Tooltip content={clicked ? "Now there's a Tooltip!" : null}>
                    <Button variant="primary" onClick={onClick}>
                        Click Me to See a Tooltip!
                    </Button>
                </Tooltip>
            </div>

            <Text>
                A Tooltip can be conditionally shown by alternating between passing <Code>null</Code> and a{' '}
                <Code>string</Code> in as <Code>content</Code>.
            </Text>
        </Grid>
    )
}

export const DefaultOpen: StoryFn = () => (
    <Grid columnCount={1}>
        <div>
            <Tooltip content="Click me!" defaultOpen={true}>
                <Button variant="primary">Example</Button>
            </Tooltip>

            <Tooltip content="Click me too!" defaultOpen={true}>
                <Button variant="primary" style={{ position: 'absolute', right: '1rem' }}>
                    Absolutely positioned example
                </Button>
            </Tooltip>
        </div>

        <Text>
            A pinned tooltip is shown on initial render (no user input required) by setting{' '}
            <Code>defaultOpen={'{true}'}</Code>.
        </Text>
    </Grid>
)

DefaultOpen.storyName = 'Default Open (Pinned)'
DefaultOpen.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

export const DisabledTrigger: StoryFn = () => (
    <Grid columnCount={1}>
        <div>
            <Tooltip content="Tooltip still works properly" placement="right">
                <Button variant="primary" disabled={true}>
                    Disabled Button 🚫
                </Button>
            </Tooltip>
        </div>

        <div>
            <Tooltip content="Tooltip still works properly" placement="right">
                <Input placeholder="Disabled Input 🚫" disabled={true} style={{ width: '300px' }} />
            </Tooltip>
        </div>

        <Text>
            Disabled <Code>{'<Button>'}</Code> and <Code>{'<Input>'}</Code> elements should work without any additional
            modifications needed.
        </Text>
    </Grid>
)

export const LongContent: StoryFn = () => (
    <Grid columnCount={1}>
        <div>
            <Tooltip
                content="Nulla porttitor accumsan tincidunt. IAmVeryLongTextWithNoBreaksAndIWantToBeWrappedInMultipleLines. Proin eget tortor risus. Quisque velit nisi, pretium ut lacinia in, elementum id enim. Donec rutrum congue leo eget malesuada."
                placement="bottom"
            >
                <Button variant="primary">Example</Button>
            </Tooltip>
        </div>

        <Text>
            Tooltips with long text will not exceed the width specified by <Code>--tooltip-max-width</Code>.
        </Text>
    </Grid>
)

export const PlacementOptions: StoryFn = () => (
    <>
        <Grid columnCount={5}>
            <div>
                <Tooltip content="Tooltip on top" placement="top">
                    <Button variant="primary">top</Button>
                </Tooltip>
            </div>

            <div>
                <Tooltip content="Tooltip on right" placement="right">
                    <Button variant="primary">right</Button>
                </Tooltip>
            </div>

            <div>
                <Tooltip content="Tooltip on bottom" placement="bottom">
                    <Button variant="primary">bottom</Button>
                </Tooltip>
            </div>

            <div>
                <Tooltip content="Tooltip on left" placement="left">
                    <Button variant="primary">left</Button>
                </Tooltip>
            </div>

            <div>
                <Tooltip content="Default Tooltip placement">
                    <Button variant="primary">(default)</Button>
                </Tooltip>
            </div>
        </Grid>

        <Text>
            The Tooltip will use the specified <Code>placement</Code> unless a viewport collision is detected, in which
            case it will be mirrored.
        </Text>
    </>
)

export const UpdateContent: StoryFn = () => {
    const [clicked, setClicked] = useState<boolean>(false)

    function onClick() {
        setClicked(true)
        setTimeout(() => setClicked(false), 1500)
    }

    return (
        <Grid columnCount={1}>
            <div>
                <Tooltip content={clicked ? 'New message!' : 'Click to change the message.'} placement="right">
                    <Button variant="primary" onClick={onClick}>
                        Click Me
                    </Button>
                </Tooltip>
            </div>

            <Text>
                The string passed in as <Code>content</Code> can be modified without any controlled or forced updates
                required.
            </Text>
        </Grid>
    )
}
