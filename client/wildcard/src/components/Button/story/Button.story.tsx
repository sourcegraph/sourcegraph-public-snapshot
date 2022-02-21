import { boolean, select } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

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

export const Simple: Story = () => (
    <Button
        variant={select('Variant', BUTTON_VARIANTS, 'primary')}
        size={select('Size', BUTTON_SIZES, undefined)}
        disabled={boolean('Disabled', false)}
        outline={boolean('Outline', false)}
    >
        Click me!
    </Button>
)

export const AllButtons: Story = () => (
    <div className="pb-3">
        <h1>Buttons</h1>
        <h2>Variants</h2>
        <ButtonVariants variants={BUTTON_VARIANTS} />
        <h2>Outline</h2>
        <ButtonVariants variants={['primary', 'secondary', 'danger']} outline={true} />
        <h2>Icons</h2>
        <p>We can use icons with our buttons.</p>
        <ButtonVariants variants={['danger']} icon={SearchIcon} />
        <ButtonVariants variants={['danger']} icon={SearchIcon} outline={true} />
        <h2>Smaller</h2>
        <p>We can make our buttons smaller.</p>
        <ButtonVariants variants={['primary']} size="sm" outline={true} />
        <h2>Links</h2>
        <p>Links can be made to look like buttons.</p>
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
        <p>Buttons can be made to look like links.</p>
        <ButtonVariants variants={['link']} />
        <h2>Button Display</h2>
        <Button className="mb-3" size="sm" variant="secondary" display="inline">
            Inline
        </Button>
        <Button size="sm" variant="secondary" display="block">
            Block
        </Button>

        <h2>Button Group</h2>
        <ButtonGroup className="mb-3">
            <Button variant="secondary" display="block">
                Grouped
            </Button>
            <Button variant="secondary" display="block">
                Grouped
            </Button>
            <Button variant="secondary" display="block">
                Grouped
            </Button>
        </ButtonGroup>
        <h2>Tooltips</h2>
        <p>Buttons can have tooltips.</p>
        <Button variant="primary" className="mr-3" data-tooltip="Some extra context on the button.">
            Enabled
        </Button>
        <Button variant="primary" disabled={true} data-tooltip="Some extra context on why the button is disabled.">
            Disabled
        </Button>
    </div>
)

AllButtons.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
