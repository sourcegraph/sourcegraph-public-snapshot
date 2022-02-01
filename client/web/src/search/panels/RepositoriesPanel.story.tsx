import { storiesOf } from '@storybook/react'
import React from 'react'
import { NEVER, of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { RepositoriesPanel } from './RepositoriesPanel'
import { _fetchRecentSearches } from './utils'

const { add } = storiesOf('web/search/panels/RepositoriesPanel', module)
    .addParameters({
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [800], disableSnapshot: false },
    })
    .addDecorator(story => <div style={{ width: '800px' }}>{story()}</div>)

const emptyRecentSearches = {
    totalCount: 0,
    nodes: [],
    pageInfo: {
        endCursor: null,
        hasNextPage: false,
    },
}

const props = {
    authenticatedUser: null,
    fetchRecentSearches: _fetchRecentSearches,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('RepositoriesPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <h2>Populated</h2>
                <RepositoriesPanel {...props} />

                <h2>Loading</h2>
                <RepositoriesPanel {...props} fetchRecentSearches={() => NEVER} />

                <h2>Empty</h2>
                <RepositoriesPanel {...props} fetchRecentSearches={() => of(emptyRecentSearches)} />
            </div>
        )}
    </WebStory>
))
