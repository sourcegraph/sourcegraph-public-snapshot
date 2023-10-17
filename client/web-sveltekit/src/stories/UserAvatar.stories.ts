import { faker } from '@faker-js/faker'
import type { Meta, StoryObj } from '@storybook/svelte'

import UserAvatar from '$lib/UserAvatar.svelte'

const meta: Meta<typeof UserAvatar> = {
    component: UserAvatar,
}

export default meta
type Story = StoryObj<typeof meta>

export const WithAvatarUrl: Story = {
    args: {
        user: {
            avatarURL: faker.internet.avatar(),
        },
    },
}

export const WithUsername: Story = {
    args: {
        user: {
            username: 'hunter',
        },
    },
}

export const WithDisplayName: Story = {
    args: {
        user: {
            displayName: 'John Doe',
        },
    },
}
