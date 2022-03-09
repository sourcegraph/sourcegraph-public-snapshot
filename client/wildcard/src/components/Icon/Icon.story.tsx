import { mdiCheck } from '@mdi/js'
import { Story, Meta } from '@storybook/react'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Icon } from './Icon'

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
        <h1>Providing SVG element</h1>
        <h3>Small Icon</h3>
        <Icon as={CheckIcon} size="sm" />
        <h3>Medium Icon</h3>
        <Icon as={CheckIcon} size="md" />
        <h1>Providing SVG path</h1>
        <h3>Small Icon</h3>
        {console.log(mdiCheck)}
        <Icon path={mdiCheck} size="sm" />
        <h3>Medium Icon</h3>
        <Icon path={mdiCheck} size="md" />
    </>
)
