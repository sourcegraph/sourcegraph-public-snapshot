import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { RouterLink } from './RouterLink'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Link/RouterLink',

    decorators: [decorator],

    parameters: {
        component: RouterLink,
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1',
        },
    },
}

export default config

export const RouterLinkExample: Story = () => (
    <p>
        Text can contain links, which <RouterLink to="/">trigger a navigation to a different page</RouterLink>.
    </p>
)
