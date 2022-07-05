import { Meta, DecoratorFn, Story } from '@storybook/react'
import { noop } from 'lodash'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { RepositoriesPanel } from './RepositoriesPanel'
import { recentSearchesPayload } from './utils'

const decorator: DecoratorFn = story => <div style={{ width: '800px' }}>{story()}</div>

const config: Meta = {
    title: 'web/search/panels/RepositoriesPanel',
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
    recentlySearchedRepositories: { recentlySearchedRepositoriesLogs: recentSearchesPayload() },
    telemetryService: NOOP_TELEMETRY_SERVICE,
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-explicit-any
    fetchMore: noop as any,
}

export const RepositoriesPanelStory: Story = () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <H2>Populated</H2>
                <RepositoriesPanel {...props} />

                <H2>Loading</H2>
                <RepositoriesPanel {...props} recentlySearchedRepositories={null} />

                <H2>Empty</H2>
                <RepositoriesPanel
                    {...props}
                    recentlySearchedRepositories={{ recentlySearchedRepositoriesLogs: emptyRecentSearches }}
                />
            </div>
        )}
    </WebStory>
)

RepositoriesPanelStory.storyName = 'RepositoriesPanel'
