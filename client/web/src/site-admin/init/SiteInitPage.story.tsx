import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../../components/WebStory'

import { SiteInitPage } from './SiteInitPage'

const config: Meta = {
    title: 'web/auth/SiteInitPage',
    parameters: {},
}

export default config

export const Default: StoryFn = () => (
    <WebStory>
        {() => (
            <SiteInitPage
                context={{
                    authMinPasswordLength: 12,
                    authPasswordPolicy: {},
                }}
                authenticatedUser={null}
                needsSiteInit={true}
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )}
    </WebStory>
)

export const Authenticated: StoryFn = () => (
    <WebStory>
        {() => (
            <SiteInitPage
                context={{
                    authMinPasswordLength: 12,
                    authPasswordPolicy: {},
                }}
                authenticatedUser={{ username: 'johndoe' }}
                needsSiteInit={true}
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )}
    </WebStory>
)
