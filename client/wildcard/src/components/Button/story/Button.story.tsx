import { useState } from 'react'

import { Meta, Story } from '@storybook/react'
import SearchIcon from 'mdi-react/SearchIcon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { H1, H2, Text, Tooltip } from '../..'
import { Button } from '../Button'
import { ButtonGroup } from '../ButtonGroup'
import { BUTTON_VARIANTS, BUTTON_SIZES } from '../constants'

import { ButtonVariants } from './ButtonVariants'

const config: Meta = {
    title: 'wildcard/Button',
    component: Button,

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

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

export const Simple: Story = (args = {}) => (
    <Button variant={args.variant} size={args.size} disabled={args.disabled} outline={args.outline}>
        Click me!
    </Button>
)
Simple.argTypes = {
    variant: {
        name: 'Variant',
        control: { type: 'select', options: BUTTON_VARIANTS },
        defaultValue: 'primary',
    },
    size: {
        name: 'Name',
        control: { type: 'select', options: BUTTON_SIZES },
        defaultValue: 'sm',
    },
    disabled: {
        name: 'Disabled',
        control: { type: 'boolean' },
        defaultValue: false,
    },
    outline: {
        name: 'Outline',
        control: { type: 'boolean' },
        defaultValue: false,
    },
}

export const AllButtons: Story = () => (
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
        <Button
            variant="secondary"
            as="a"
            href="https://example.com"
            target="_blank"
            rel="noopener noreferrer"
            className="mb-3"
        >
            I am a link
        </Button>
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

AllButtons.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

export const Group: Story = () => {
    const [selectedButton, setSelectedButton] = useState(1)

    return (
        <>
            <div className="mb-4">
                <H2>Standard</H2>

                <ButtonGroup>
                    <Button variant="secondary" outline={selectedButton !== 1} onClick={() => setSelectedButton(1)}>
                        One
                    </Button>
                    <Button variant="secondary" outline={selectedButton !== 2} onClick={() => setSelectedButton(2)}>
                        Two
                    </Button>
                    <Button variant="secondary" outline={selectedButton !== 3} onClick={() => setSelectedButton(3)}>
                        Three
                    </Button>
                </ButtonGroup>
            </div>

            <div className="mb-4">
                <H2>With Tooltips</H2>

                <ButtonGroup>
                    <Tooltip content="Option One">
                        <Button variant="secondary" outline={selectedButton !== 1} onClick={() => setSelectedButton(1)}>
                            One
                        </Button>
                    </Tooltip>
                    <Tooltip content="Option Two">
                        <Button variant="secondary" outline={selectedButton !== 2} onClick={() => setSelectedButton(2)}>
                            Two
                        </Button>
                    </Tooltip>
                    <Tooltip content="Option Three">
                        <Button variant="secondary" outline={selectedButton !== 3} onClick={() => setSelectedButton(3)}>
                            Three
                        </Button>
                    </Tooltip>
                </ButtonGroup>
            </div>

            <div className="mb-4">
                <H2>With Tooltips (Disabled Buttons)</H2>

                <ButtonGroup>
                    <Tooltip content="Option One">
                        <Button variant="secondary" outline={selectedButton !== 1} onClick={() => setSelectedButton(1)}>
                            One
                        </Button>
                    </Tooltip>
                    <Tooltip content="Option Two">
                        <Button
                            disabled={true}
                            variant="secondary"
                            outline={selectedButton !== 2}
                            onClick={() => setSelectedButton(2)}
                        >
                            Two
                        </Button>
                    </Tooltip>
                    <Tooltip content="Option Three">
                        <Button
                            disabled={true}
                            variant="secondary"
                            outline={selectedButton !== 3}
                            onClick={() => setSelectedButton(3)}
                        >
                            Three
                        </Button>
                    </Tooltip>
                </ButtonGroup>
            </div>
        </>
    )
}
Group.storyName = 'Button Group'
Group.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
