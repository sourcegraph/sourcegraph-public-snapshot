import { Meta, DecoratorFn, Story } from '@storybook/react'
import { parseISO } from 'date-fns'
import { noop } from 'lodash'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { RecentSearchesPanel } from './RecentSearchesPanel'
import { recentSearchesPayload } from './utils'

const decorator: DecoratorFn = story => <div style={{ width: '800px' }}>{story()}</div>

const config: Meta = {
    title: 'web/search/panels/RecentSearchesPanel',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [800], disableSnapshot: false },
    },
    decorators: [decorator],
}

export default config

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

export const RecentSearchesPanelStory: Story = () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <H2>Populated</H2>
                <RecentSearchesPanel {...props} />

                <H2>Loading</H2>
                <RecentSearchesPanel {...props} recentSearches={null} />

                <H2>Empty</H2>
                <RecentSearchesPanel {...props} recentSearches={{ recentSearchesLogs: emptyRecentSearches }} />
            </div>
        )}
    </WebStory>
)

RecentSearchesPanelStory.storyName = 'RecentSearchesPanel'
