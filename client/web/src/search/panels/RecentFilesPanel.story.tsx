import React from 'react'

import { storiesOf } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

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
    loadMore: () => {},
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('RecentFilesPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <h2>Populated</h2>
                <RecentFilesPanel {...props} />

                <h2>Loading</h2>
                <RecentFilesPanel {...props} recentFilesFragment={null} />

                <h2>Empty</h2>
                <RecentFilesPanel
                    {...props}
                    recentFilesFragment={{
                        recentFilesLogs: {
                            nodes: [],
                            totalCount: 0,
                            pageInfo: { hasNextPage: false },
                        },
                    }}
                />
            </div>
        )}
    </WebStory>
))
