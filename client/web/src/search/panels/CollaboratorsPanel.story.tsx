import { storiesOf } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { CollaboratorsPanel } from './CollaboratorsPanel'
import { collaboratorsPayload, authUser } from './utils'

const { add } = storiesOf('web/search/panels/CollaboratorsPanel', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/Og1zVk7BbZ7SWTXM5WsWA5/Account-Setups-OKR-explorations?node-id=188%3A17448',
    },
    chromatic: { disableSnapshot: false },
})

const props = {
    authenticatedUser: authUser,
    collaboratorsFragment: { collaborators: collaboratorsPayload() },
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('CollaboratorsPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <Typography.H2>Populated</Typography.H2>
                <CollaboratorsPanel {...props} />

                <Typography.H2>Loading</Typography.H2>
                <CollaboratorsPanel {...props} collaboratorsFragment={null} />

                <Typography.H2>Empty</Typography.H2>
                <CollaboratorsPanel {...props} collaboratorsFragment={{ collaborators: [] }} />
            </div>
        )}
    </WebStory>
))
