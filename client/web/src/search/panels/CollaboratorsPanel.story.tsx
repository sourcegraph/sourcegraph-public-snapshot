import { storiesOf } from '@storybook/react'
import React from 'react'
import { NEVER, of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { CollaboratorsPanel } from './CollaboratorsPanel'
import { _fetchCollaborators, authUser } from './utils'

const { add } = storiesOf('web/search/panels/CollaboratorsPanel', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/Og1zVk7BbZ7SWTXM5WsWA5/Account-Setups-OKR-explorations?node-id=188%3A17448',
    },
    chromatic: { disableSnapshot: false },
})

const props = {
    authenticatedUser: authUser,
    fetchCollaborators: _fetchCollaborators,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('CollaboratorsPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <h2>Populated</h2>
                <CollaboratorsPanel {...props} />

                <h2>Loading</h2>
                <CollaboratorsPanel {...props} fetchCollaborators={() => NEVER} />

                <h2>Empty</h2>
                <CollaboratorsPanel {...props} fetchCollaborators={() => of([])} />
            </div>
        )}
    </WebStory>
))
