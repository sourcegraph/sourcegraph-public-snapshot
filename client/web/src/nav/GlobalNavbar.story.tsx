import { storiesOf } from '@storybook/react'
import { createMemoryHistory } from 'history'
import { SuiteFunction } from 'mocha'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'
import { SourcegraphContext } from '../jscontext'
import { useExperimentalFeatures } from '../stores'
import { ThemePreference } from '../stores/themeState'

import { GlobalNavbar } from './GlobalNavbar'

if (!window.context) {
    window.context = {} as SourcegraphContext & SuiteFunction
}

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
    keyboardShortcuts: [],
    isLightTheme: props.isLightTheme,
    isExtensionAlertAnimating: false,
    batchChangesEnabled: true,
    batchChangesExecutionEnabled: true,
    batchChangesWebhookLogsEnabled: true,
    activation: undefined,
    routes: [],
    extensionViews: () => null,
})

const { add } = storiesOf('web/nav/GlobalNav', module).addDecorator(Story => {
    useExperimentalFeatures.setState({ codeMonitoring: true })
    return <Story />
})

add('Anonymous viewer', () => (
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
))

add('Auth required', () => (
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
))

add(
    'Authenticated viewer',
    () => (
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
    ),
    {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/SFhXbl23TJ2j5tOF51NDtF/%F0%9F%93%9AWeb?node-id=985%3A1281',
        },
    }
)
