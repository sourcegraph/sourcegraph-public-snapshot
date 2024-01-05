import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../components/WebStory'
import type { UserAreaUserFields } from '../../graphql-operations'

import { UserProfile } from './UserProfile'

const decorator: Decorator = story => <div className="p-3 container list-unstyled">{story()}</div>

const config: Meta = {
    title: 'web/src/user/profile',
    decorators: [decorator],
}

export default config

const mockUser: UserAreaUserFields = {
    __typename: 'User',
    id: 'test-id',
    username: 'alice',
    displayName: 'alice',
    url: '/user/alice',
    settingsURL: null,
    avatarURL: null,
    viewerCanAdminister: true,
    builtinAuth: true,
    createdAt: '2023-03-02 16:34:05.882711+01',
    emails: [
        {
            __typename: 'UserEmail',
            email: 'alice@test.com',
            isPrimary: true,
        },
        {
            __typename: 'UserEmail',
            email: 'alice2@test.com',
            isPrimary: false,
        },
    ],
    roles: {
        __typename: 'RoleConnection',
        nodes: [
            { name: 'USER', system: true },
            { name: 'Operator', system: false },
        ],
    },
}

export const Profile: StoryFn = () => <WebStory>{() => <UserProfile user={mockUser} />}</WebStory>
