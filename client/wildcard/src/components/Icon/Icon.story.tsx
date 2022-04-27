import React from 'react'

import { mdiCheck } from '@mdi/js'
import { Story, Meta } from '@storybook/react'
import CheckIcon from 'mdi-react/CheckIcon'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { H1, H2, H3, Code } from '..'

import { IconStyle, Icon } from '.'

const config: Meta = {
    title: 'wildcard/Icon',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Icon,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=1366%3A611',
        },
    },
}
export default config

export const Simple: Story = () => (
    <>
        <H1>Icons</H1>
        <H2>We can render different elements as Icons</H2>
        <H3>Small</H3>
        <IconStyle size="sm" as="img" src="https://picsum.photos/id/237/50" alt="A dog" />
        <h3>Medium</h3>
        <IconStyle size="md" as="img" src="https://picsum.photos/id/237/50" alt="A dog" />

        <H2 className="mt-3">We can render SVGs as Icons</H2>
        <p>
            Note: This is typically used for our older <Code>mdi-react</Code> SVG components.
        </p>
        <H3>Small</H3>
        <Icon as={CheckIcon} size="sm" />
        <H3>Medium</H3>
        <Icon as={CheckIcon} size="md" />

        <H2 className="mt-3">We can render SVG paths as Icons</H2>
        <p>
            Note: This is typically used for our newer <Code>@mdi/react</Code> SVGs. This method allows us to inject
            additional elements required for accessibility, such as <Code>{'<title>'}</Code>.
        </p>
        <H3>Small</H3>
        <Icon svgPath={mdiCheck} size="sm" title="Checkmark" />
        <H3>Medium</H3>
        <Icon svgPath={mdiCheck} size="md" title="Checkmark" />
    </>
)
