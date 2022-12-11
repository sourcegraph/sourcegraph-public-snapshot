import { Story, Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { CollaboratorsPanel } from './CollaboratorsPanel'
import { collaboratorsPayload, authUser } from './utils'

const config: Meta = {
    title: 'web/search/panels/CollaboratorsPanel',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/Og1zVk7BbZ7SWTXM5WsWA5/Account-Setups-OKR-explorations?node-id=188%3A17448',
        },
        chromatic: { disableSnapshot: false },
    },
}

export default config

const props = {
    authenticatedUser: authUser,
    collaboratorsFragment: { collaborators: collaboratorsPayload() },
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

export const CollaboratorsPanelStory: Story = () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <H2>Populated</H2>
                <CollaboratorsPanel {...props} />

                <H2>Loading</H2>
                <CollaboratorsPanel {...props} collaboratorsFragment={null} />

                <H2>Empty</H2>
                <CollaboratorsPanel {...props} collaboratorsFragment={{ collaborators: [] }} />
            </div>
        )}
    </WebStory>
)
CollaboratorsPanelStory.storyName = 'CollaboratorsPanel'
