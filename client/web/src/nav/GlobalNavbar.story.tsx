import React from 'react'

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

import { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'
import { useExperimentalFeatures } from '../stores'
import { ThemePreference } from '../stores/themeState'

import { GlobalNavbar } from './GlobalNavbar'

const history = createMemoryHistory()

const defaultProps = (
    props: ThemeProps
): Omit<
    React.ComponentProps<typeof GlobalNavbar>,
    'authenticatedUser' | 'variant' | 'showSearchBox' | 'authRequired'
> => ({
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
    searchContextsEnabled: true,
    batchChangesEnabled: true,
    batchChangesExecutionEnabled: true,
    batchChangesWebhookLogsEnabled: true,
    activation: undefined,
    routes: [],
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    showKeyboardShortcutsHelp: () => undefined,
})

const decorator: DecoratorFn = Story => {
    useExperimentalFeatures.setState({ codeMonitoring: true })
    return <Story />
}

const config: Meta = {
    title: 'web/nav/GlobalNav',
    decorators: [decorator],
}

export default config

export const AnonymousViewer: Story = () => (
    <WebStory>
        {webProps => (
            <GlobalNavbar
                {...defaultProps(webProps)}
                authRequired={false}
                authenticatedUser={null}
                variant="default"
                showSearchBox={false}
            />
        )}
    </WebStory>
)

AnonymousViewer.storyName = 'Anonymous viewer'

export const AuthRequired: Story = () => (
    <WebStory>
        {webProps => (
            <GlobalNavbar
                {...defaultProps(webProps)}
                authRequired={true}
                authenticatedUser={null}
                variant="default"
                showSearchBox={false}
            />
        )}
    </WebStory>
)

AuthRequired.storyName = 'Auth required'

export const AuthenticatedViewer: Story = () => (
    <WebStory>
        {webProps => (
            <GlobalNavbar
                {...defaultProps(webProps)}
                authRequired={false}
                authenticatedUser={
                    { username: 'alice', organizations: { nodes: [{ name: 'acme' }] } } as AuthenticatedUser
                }
                variant="default"
                showSearchBox={false}
            />
        )}
    </WebStory>
)

AuthenticatedViewer.storyName = 'Authenticated viewer'

AuthenticatedViewer.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/SFhXbl23TJ2j5tOF51NDtF/%F0%9F%93%9AWeb?node-id=985%3A1281',
    },
}
