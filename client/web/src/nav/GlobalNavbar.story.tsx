import { DecoratorFn, Meta, Story } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { Grid, H3 } from '@sourcegraph/wildcard'

import { HistoryStack } from '../app/useHistoryStack'
import { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'

import { GlobalNavbar, GlobalNavbarProps } from './GlobalNavbar'

const defaultProps: GlobalNavbarProps = {
    isSourcegraphDotCom: false,
    isSourcegraphApp: false,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    telemetryService: NOOP_TELEMETRY_SERVICE,
    platformContext: {} as any,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => undefined,
    searchContextsEnabled: false,
    batchChangesEnabled: false,
    batchChangesExecutionEnabled: false,
    batchChangesWebhookLogsEnabled: false,
    routes: [],
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    showKeyboardShortcutsHelp: () => undefined,
    showSearchBox: false,
    authenticatedUser: null,
    setFuzzyFinderIsVisible: () => {},
    notebooksEnabled: true,
    codeMonitoringEnabled: true,
    ownEnabled: true,
    showFeedbackModal: () => undefined,
    historyStack: {} as HistoryStack,
}

const allNavItemsProps: Partial<GlobalNavbarProps> = {
    searchContextsEnabled: true,
    batchChangesEnabled: true,
    batchChangesExecutionEnabled: true,
    batchChangesWebhookLogsEnabled: true,
    codeInsightsEnabled: true,
}

const allAuthenticatedNavItemsProps: Partial<GlobalNavbarProps> = {
    authenticatedUser: {
        username: 'alice',
        organizations: { nodes: [{ id: 'acme', name: 'acme' }] },
        siteAdmin: true,
    } as AuthenticatedUser,
}

const decorator: DecoratorFn = Story => (
    <WebStory>
        {() => (
            <div className="mt-3">
                <Story args={defaultProps} />
            </div>
        )}
    </WebStory>
)

const config: Meta = {
    title: 'web/nav/GlobalNav',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
            viewports: [320, 576, 978],
        },
    },
}

export default config

export const Default: Story<GlobalNavbarProps> = props => (
    <Grid columnCount={1}>
        <div>
            <H3 className="ml-2">Anonymous viewer</H3>
            <GlobalNavbar {...props} />
        </div>
        <div>
            <H3 className="ml-2">Anonymous viewer with all possible nav items</H3>
            <GlobalNavbar {...props} {...allNavItemsProps} />
        </div>
        <div>
            <H3 className="ml-2">Authenticated user with all possible nav items</H3>
            <GlobalNavbar {...props} {...allNavItemsProps} {...allAuthenticatedNavItemsProps} />
        </div>
        <div>
            <H3 className="ml-2">Authenticated user with all possible nav items and search input</H3>
            <GlobalNavbar {...props} {...allNavItemsProps} {...allAuthenticatedNavItemsProps} showSearchBox={true} />
        </div>
    </Grid>
)
