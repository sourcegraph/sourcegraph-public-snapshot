import { boolean, select } from '@storybook/addon-knobs'
import { Meta, Story } from '@storybook/react'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../Button'
import { BUTTON_VARIANTS, BUTTON_SIZES } from '../constants'

import { ButtonVariants } from './ButtonVariants'

const config: Meta = {
    title: 'wildcard/Button',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Button,
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A2514',
        },
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
    <>
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
    </>
)
