import { useState } from 'react'

import type { Meta, StoryFn } from '@storybook/react'
import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'

import { H1, H2, Text, Tooltip, ButtonLink, Code } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'
import { Button } from '../Button'
import { ButtonGroup } from '../ButtonGroup'
import { BUTTON_VARIANTS, BUTTON_SIZES } from '../constants'

import { ButtonVariants } from './ButtonVariants'

const config: Meta = {
    title: 'wildcard/Button',
    component: Button,

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: Button,
        design: [
            {
                type: 'figma',
                name: 'Figma Light',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A2513',
            },
            {
                type: 'figma',
                name: 'Figma Dark',
                url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A5794',
            },
        ],
    },
}

export default config

export const Simple: StoryFn = (args = {}) => (
    <Button variant={args.variant} size={args.size} disabled={args.disabled} outline={args.outline}>
        Click me!
    </Button>
)
Simple.argTypes = {
    variant: {
        name: 'Variant',
        control: { type: 'select', options: BUTTON_VARIANTS },
    },
    size: {
        name: 'Name',
        control: { type: 'select', options: BUTTON_SIZES },
    },
    disabled: {
        name: 'Disabled',
        control: { type: 'boolean' },
    },
    outline: {
        name: 'Outline',
        control: { type: 'boolean' },
    },
}
Simple.args = {
    variant: 'primary',
    size: 'sm',
    disabled: false,
    outline: false,
}

export const AllButtons: StoryFn = () => (
    <div className="pb-3">
        <H1>Buttons</H1>
        <H2>Variants</H2>
        <ButtonVariants variants={BUTTON_VARIANTS} />
        <H2>Outline</H2>
        <ButtonVariants variants={['primary', 'secondary', 'danger']} outline={true} />
        <H2>Icons</H2>
        <Text>We can use icons with our buttons.</Text>
        <ButtonVariants variants={['danger']} icon={SearchIcon} />
        <ButtonVariants variants={['danger']} icon={SearchIcon} outline={true} />
        <H2>Smaller</H2>
        <Text>We can make our buttons smaller.</Text>
        <ButtonVariants variants={['primary']} size="sm" outline={true} />
        <H2>Links</H2>
        <Text>Links can be made to look like buttons.</Text>
        <ButtonLink
            variant="secondary"
            to="https://example.com"
            target="_blank"
            rel="noopener noreferrer"
            className="mb-3"
        >
            I am a link
        </ButtonLink>
        <Text>Buttons can be made to look like links.</Text>
        <ButtonVariants variants={['link']} />
        <H2>Button Display</H2>
        <Button className="mb-3" size="sm" variant="secondary" display="inline">
            Inline
        </Button>
        <Button size="sm" variant="secondary" display="block">
            Block
        </Button>

        <H2>Tooltips</H2>
        <Text>Buttons can have tooltips.</Text>
        <Tooltip content="Some extra context on the button.">
            <Button variant="primary" className="mr-3">
                Enabled
            </Button>
        </Tooltip>
        <Tooltip content="Some extra context on why the button is disabled.">
            <Button variant="primary" disabled={true}>
                Disabled
            </Button>
        </Tooltip>
    </div>
)

type ButtonSizesType = typeof BUTTON_SIZES[number] | undefined

export const Group: StoryFn = () => {
    const [active, setActive] = useState<'Left' | 'Middle' | 'Right'>('Left')
    const buttonSizes: ButtonSizesType[] = ['lg', undefined, 'sm']

    return (
        <>
            <H1>Button groups</H1>

            <H2>Example</H2>
            <div className="mb-2">
                <Text>
                    Button groups have no styles on their own, they just group buttons together. This means they can be
                    used to group any other semantic or outline button variant.
                </Text>
                <div className="mb-2">
                    <ButtonGroup aria-label="Basic example">
                        <Button variant="secondary">Left</Button>
                        <Button variant="secondary">Middle</Button>
                        <Button variant="secondary">Right</Button>
                    </ButtonGroup>{' '}
                    Example with secondary buttons
                </div>
                <div className="mb-2">
                    <ButtonGroup aria-label="Basic example">
                        <Button outline={true} variant="secondary">
                            Left
                        </Button>
                        <Button outline={true} variant="secondary">
                            Middle
                        </Button>
                        <Button outline={true} variant="secondary">
                            Right
                        </Button>
                    </ButtonGroup>{' '}
                    Example with secondary outline buttons
                </div>
                <div className="mb-2">
                    <ButtonGroup aria-label="Basic example">
                        <Button outline={true} variant="primary">
                            Left
                        </Button>
                        <Button outline={true} variant="primary">
                            Middle
                        </Button>
                        <Button outline={true} variant="primary">
                            Right
                        </Button>
                    </ButtonGroup>{' '}
                    Example with primary outline buttons
                </div>
            </div>

            <H2 className="mt-3">Sizing</H2>
            <Text>
                Just like buttons, button groups have <Code>sm</Code> and <Code>lg</Code> size variants.
            </Text>
            <div className="mb-2">
                {buttonSizes.map(size => (
                    <div key={size} className="mb-2">
                        <ButtonGroup aria-label="Sizing example">
                            <Button size={size} outline={true} variant="primary">
                                Left
                            </Button>
                            <Button size={size} outline={true} variant="primary">
                                Middle
                            </Button>
                            <Button size={size} outline={true} variant="primary">
                                Right
                            </Button>
                        </ButtonGroup>
                    </div>
                ))}
            </div>

            <H2 className="mt-3">Active state</H2>
            <Text>
                The <Code>active</Code> class can be used to craft toggles out of button groups.
            </Text>
            <div className="mb-2">
                <ButtonGroup aria-label="Basic example">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Button
                            key={option}
                            className={classNames(option === active && 'active')}
                            onClick={() => setActive(option)}
                            aria-pressed={option === active}
                            outline={true}
                            variant="secondary"
                        >
                            {option}
                        </Button>
                    ))}
                </ButtonGroup>{' '}
                Example with secondary outline buttons
            </div>
            <div className="mb-2">
                <ButtonGroup aria-label="Basic example">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Button
                            key={option}
                            className={classNames(option === active && 'active')}
                            onClick={() => setActive(option)}
                            aria-pressed={option === active}
                            outline={true}
                            variant="primary"
                        >
                            {option}
                        </Button>
                    ))}
                </ButtonGroup>{' '}
                Example with primary outline buttons
            </div>
            <div className="mb-2">
                <ButtonGroup aria-label="Basic example">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Button
                            key={option}
                            className={classNames(option === active && 'active')}
                            onClick={() => setActive(option)}
                            aria-pressed={option === active}
                            variant="secondary"
                        >
                            {option}
                        </Button>
                    ))}
                </ButtonGroup>{' '}
                Example with secondary buttons
            </div>
            <div className="mb-2">
                <ButtonGroup aria-label="Basic example">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Button
                            key={option}
                            className={classNames(option === active && 'active')}
                            onClick={() => setActive(option)}
                            aria-pressed={option === active}
                            variant="primary"
                        >
                            {option}
                        </Button>
                    ))}
                </ButtonGroup>{' '}
                Example with primary buttons
            </div>
            <div className="mb-2">
                <ButtonGroup aria-label="Basic example">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Button
                            key={option}
                            className={classNames(option === active && 'active')}
                            onClick={() => setActive(option)}
                            aria-pressed={option === active}
                            variant="link"
                        >
                            {option}
                        </Button>
                    ))}
                </ButtonGroup>{' '}
                Example with link buttons
            </div>

            <H2 className="mt-3">With Tooltips</H2>
            <div className="mb-2">
                <ButtonGroup aria-label="With Tooltips">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Tooltip key={option} content={`Option ${option}`}>
                            <Button
                                variant="secondary"
                                outline={option === active}
                                onClick={() => setActive(option)}
                                aria-pressed={option === active}
                            >
                                {option}
                            </Button>
                        </Tooltip>
                    ))}
                </ButtonGroup>{' '}
                Example with enabled buttons
            </div>
            <div className="mb-2">
                <ButtonGroup aria-label="With Tooltips (Disabled Buttons)">
                    {(['Left', 'Middle', 'Right'] as const).map(option => (
                        <Tooltip key={option} content={`Option ${option}`}>
                            <Button
                                variant="secondary"
                                disabled={true}
                                outline={option === active}
                                onClick={() => setActive(option)}
                                aria-pressed={option === active}
                            >
                                {option}
                            </Button>
                        </Tooltip>
                    ))}
                </ButtonGroup>{' '}
                Example with disabled buttons
            </div>
        </>
    )
}

Group.storyName = 'Button Group'
