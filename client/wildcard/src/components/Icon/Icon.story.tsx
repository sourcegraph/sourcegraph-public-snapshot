import { Story, Meta } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Typography } from '..'
import { SourcegraphIcon } from '../SourcegraphIcon'

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
        <Typography.H3>Small Icon</Typography.H3>
        <Icon as={SourcegraphIcon} size="sm" title="Sourcegraph logo" />

        <Typography.H3>Medium Icon</Typography.H3>
        <Icon as={SourcegraphIcon} size="md" aria-label="Sourcegraph logo" />
    </>
)
