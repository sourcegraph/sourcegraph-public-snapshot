import { storiesOf } from '@storybook/react'
import { parseISO } from 'date-fns'
import { noop } from 'lodash'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { RecentSearchesPanel } from './RecentSearchesPanel'
import { recentSearchesPayload } from './utils'

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
    recentSearches: { recentSearchesLogs: recentSearchesPayload() },
    now: () => parseISO('2020-09-16T23:15:01Z'),
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
    fetchMore: noop as any,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('RecentSearchesPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <Typography.H2>Populated</Typography.H2>
                <RecentSearchesPanel {...props} />

                <Typography.H2>Loading</Typography.H2>
                <RecentSearchesPanel {...props} recentSearches={null} />

                <Typography.H2>Empty</Typography.H2>
                <RecentSearchesPanel {...props} recentSearches={{ recentSearchesLogs: emptyRecentSearches }} />
            </div>
        )}
    </WebStory>
))
