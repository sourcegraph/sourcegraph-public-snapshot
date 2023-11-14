import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import type { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'

import { PostSignUpPage } from './PostSignUpPage'

const config: Meta = {
    title: 'web/auth/PostSignUpPage',
}

export default config

const mockUser = {
    __typename: 'User',
    id: '1',
    username: 'user',
    emails: [{ email: 'user@me.com', isPrimary: true, verified: false }],
    hasVerifiedEmail: false,
    completedPostSignup: false,
} as AuthenticatedUser

const telemetryRecorder = noOpTelemetryRecorder

export const UnverifiedEmail: StoryFn = () => (
    <WebStory>{() => <PostSignUpPage authenticatedUser={mockUser} telemetryRecorder={telemetryRecorder} />}</WebStory>
)

export const VerifiedEmail: StoryFn = () => (
    <WebStory>
        {() => (
            <PostSignUpPage
                authenticatedUser={{ ...mockUser, hasVerifiedEmail: true }}
                telemetryRecorder={telemetryRecorder}
            />
        )}
    </WebStory>
)
