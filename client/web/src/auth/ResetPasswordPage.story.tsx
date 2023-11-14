import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import type { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'

import { ResetPasswordPage } from './ResetPasswordPage'

const config: Meta = {
    title: 'web/auth/ResetPasswordPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: StoryFn = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                telemetryRecorder={noOpTelemetryRecorder}
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const WithCode: StoryFn = () => (
    <WebStory initialEntries={[{ pathname: '/reset-password', search: '?code=123123&userID=123' }]}>
        {() => (
            <ResetPasswordPage
                telemetryRecorder={noOpTelemetryRecorder}
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const LoggedInUser: StoryFn = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                telemetryRecorder={noOpTelemetryRecorder}
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: false }}
                authenticatedUser={{ id: 'user' } as AuthenticatedUser}
            />
        )}
    </WebStory>
)

export const Disabled: StoryFn = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                telemetryRecorder={noOpTelemetryRecorder}
                context={{ xhrHeaders: {}, resetPasswordEnabled: false, sourcegraphDotComMode: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const Dotcom: StoryFn = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                telemetryRecorder={noOpTelemetryRecorder}
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: true }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)
