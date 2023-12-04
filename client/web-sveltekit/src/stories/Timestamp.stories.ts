import type { Meta, StoryObj } from '@storybook/svelte'

import TimestampExample from './TimestampExample.svelte'

const meta: Meta<typeof TimestampExample> = {
    component: TimestampExample,
}

export default meta
type Story = StoryObj<typeof meta>

export const Timestamp: Story = {}
