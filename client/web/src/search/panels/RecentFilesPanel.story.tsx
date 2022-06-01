import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { RecentFilesPanel } from './RecentFilesPanel'
import { recentFilesPayload } from './utils'

const { add } = storiesOf('web/search/panels/RecentFilesPanel', module)
    .addParameters({
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/sPRyyv3nt5h0284nqEuAXE/12192-Sourcegraph-server-page-v1?node-id=255%3A3',
        },
        chromatic: { viewports: [800], disableSnapshot: false },
    })
    .addDecorator(story => <div style={{ width: '800px' }}>{story()}</div>)

const emptyRecentFiles = {
    totalCount: 0,
    nodes: [],
    pageInfo: {
        endCursor: null,
        hasNextPage: false,
    },
}

const props = {
    authenticatedUser: null,
    recentFilesFragment: { recentFilesLogs: recentFilesPayload() },
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
    fetchMore: noop as any,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('RecentFilesPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <Typography.H2>Populated</Typography.H2>
                <RecentFilesPanel {...props} />

                <Typography.H2>Loading</Typography.H2>
                <RecentFilesPanel {...props} recentFilesFragment={null} />

                <Typography.H2>Empty</Typography.H2>
                <RecentFilesPanel {...props} recentFilesFragment={{ recentFilesLogs: emptyRecentFiles }} />
            </div>
        )}
    </WebStory>
))
