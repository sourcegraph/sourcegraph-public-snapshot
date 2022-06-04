import { storiesOf } from '@storybook/react'

import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { SavedSearchesPanel } from './SavedSearchesPanel'
import { savedSearchesPayload, authUser } from './utils'

const { add } = storiesOf('web/search/panels/SavedSearchesPanel', module).addParameters({
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
    },
    chromatic: { disableSnapshot: false },
})

const props = {
    authenticatedUser: authUser,
    patternType: SearchPatternType.literal,
    savedSearchesFragment: { savedSearches: savedSearchesPayload() },
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('SavedSearchesPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <H2>Populated</H2>
                <SavedSearchesPanel {...props} />

                <H2>Loading</H2>
                <SavedSearchesPanel {...props} savedSearchesFragment={null} />

                <H2>Empty</H2>
                <SavedSearchesPanel {...props} savedSearchesFragment={{ savedSearches: [] }} />
            </div>
        )}
    </WebStory>
))
