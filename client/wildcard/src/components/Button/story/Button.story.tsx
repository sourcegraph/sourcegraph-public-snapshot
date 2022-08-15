import { boolean, select } from '@storybook/addon-knobs'
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

        <H2>Button Group</H2>
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
