import type { Meta, StoryFn } from '@storybook/react'

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
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: true }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)
