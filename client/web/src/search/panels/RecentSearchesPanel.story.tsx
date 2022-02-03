import { storiesOf } from '@storybook/react'
import { parseISO } from 'date-fns'
import React from 'react'
import { NEVER, of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { RecentSearchesPanel } from './RecentSearchesPanel'
import { _fetchRecentSearches } from './utils'

const { add } = storiesOf('web/search/panels/RecentSearchesPanel', module)
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
    now: () => parseISO('2020-09-16T23:15:01Z'),
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('RecentSearchesPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <h2>Populated</h2>
                <RecentSearchesPanel {...props} />

                <h2>Loading</h2>
                <RecentSearchesPanel {...props} fetchRecentSearches={() => NEVER} />

                <h2>Empty</h2>
                <RecentSearchesPanel {...props} fetchRecentSearches={() => of(emptyRecentSearches)} />
            </div>
        )}
    </WebStory>
))
