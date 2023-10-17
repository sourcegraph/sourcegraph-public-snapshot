import type { Meta, StoryFn } from '@storybook/react'

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

export const UnverifiedEmail: StoryFn = () => (
    <WebStory>{() => <PostSignUpPage authenticatedUser={mockUser} />}</WebStory>
)

export const VerifiedEmail: StoryFn = () => (
    <WebStory>{() => <PostSignUpPage authenticatedUser={{ ...mockUser, hasVerifiedEmail: true }} />}</WebStory>
)
