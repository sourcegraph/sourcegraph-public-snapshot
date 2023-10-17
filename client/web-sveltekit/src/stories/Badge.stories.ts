import type { Meta, StoryObj } from '@storybook/svelte'

import BadgeExample from './BadgeExample.svelte'

const meta: Meta<typeof BadgeExample> = {
    title: 'wildcard/Badge',
    component: BadgeExample,
}

export default meta
type Story = StoryObj<typeof meta>

export const Badge: Story = {}
