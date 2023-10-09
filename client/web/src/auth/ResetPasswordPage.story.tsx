import type { Meta, Story } from '@storybook/react'

import { AuthenticatedUser } from '../auth'
import { WebStory } from '../components/WebStory'

import { ResetPasswordPage } from './ResetPasswordPage'

const config: Meta = {
    title: 'web/auth/ResetPasswordPage',
}

export default config

export const Default: Story = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const WithCode: Story = () => (
    <WebStory initialEntries={[{ pathname: '/reset-password', search: '?code=123123&userID=123' }]}>
        {() => (
            <ResetPasswordPage
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const LoggedInUser: Story = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: false }}
                authenticatedUser={{ id: 'user' } as AuthenticatedUser}
            />
        )}
    </WebStory>
)

export const Disabled: Story = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                context={{ xhrHeaders: {}, resetPasswordEnabled: false, sourcegraphDotComMode: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const Dotcom: Story = () => (
    <WebStory>
        {() => (
            <ResetPasswordPage
                context={{ xhrHeaders: {}, resetPasswordEnabled: true, sourcegraphDotComMode: true }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)
