import { DecoratorFn, Meta, Story } from '@storybook/react'
import { createMemoryHistory } from 'history'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Grid, H3 } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { baseActivation } from '../components/ActivationDropdown/ActivationDropdown.fixtures'
import { WebStory } from '../components/WebStory'
import { useExperimentalFeatures } from '../stores'
import { ThemePreference } from '../theme'

import { GlobalNavbar, GlobalNavbarProps } from './GlobalNavbar'

const history = createMemoryHistory()

const getDefaultProps = (props: ThemeProps): GlobalNavbarProps => ({
    isSourcegraphDotCom: false,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    location: history.location,
    history,
    extensionsController,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    themePreference: ThemePreference.Light,
    onThemePreferenceChange: () => undefined,
    globbing: false,
    platformContext: {} as any,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => undefined,
    defaultSearchContextSpec: '',
    isLightTheme: props.isLightTheme,
    searchContextsEnabled: false,
    batchChangesEnabled: false,
    batchChangesExecutionEnabled: false,
    batchChangesWebhookLogsEnabled: false,
    activation: undefined,
    routes: [],
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    showKeyboardShortcutsHelp: () => undefined,
    showSearchBox: false,
    authenticatedUser: null,
    setFuzzyFinderIsVisible: () => {},
})

const allNavItemsProps: Partial<GlobalNavbarProps> = {
    searchContextsEnabled: true,
    batchChangesEnabled: true,
    batchChangesExecutionEnabled: true,
    batchChangesWebhookLogsEnabled: true,
    codeInsightsEnabled: true,
    enableLegacyExtensions: true,
}

const allAuthenticatedNavItemsProps: Partial<GlobalNavbarProps> = {
    activation: { ...baseActivation(), completed: { ConnectedCodeHost: true, DidSearch: false } },
    authenticatedUser: {
        username: 'alice',
        organizations: { nodes: [{ id: 'acme', name: 'acme' }] },
        siteAdmin: true,
    } as AuthenticatedUser,
}

const decorator: DecoratorFn = Story => {
    useExperimentalFeatures.setState({ codeMonitoring: true })

    return (
        <WebStory>
            {props => (
                <div className="mt-3">
                    <Story args={getDefaultProps(props)} />
                </div>
            )}
        </WebStory>
    )
}

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
