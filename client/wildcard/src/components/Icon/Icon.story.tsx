import { mdiClose } from '@mdi/js'
import { Story, Meta } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { IconV2 } from '..'
import { H3 } from '../..'
import { SourcegraphIcon } from '../SourcegraphIcon'

const config: Meta = {
    title: 'wildcard/Icon',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: IconV2,
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
        <H3>Small Icon</H3>
        <IconV2 as={SourcegraphIcon} size="sm" aria-label="Sourcegraph logo" />

        <H3>Medium Icon</H3>
        <IconV2 as={SourcegraphIcon} size="md" aria-label="Sourcegraph logo" />

        <H3>MDI Icon</H3>
        <IconV2 svgPath={mdiClose} size="md" aria-label="Close" />
    </>
)
