import type { Meta, StoryObj } from '@storybook/svelte'

import SeparatorExample from './SeparatorExample.svelte'

const meta: Meta<typeof SeparatorExample> = {
    component: SeparatorExample,
}

export default meta
type Story = StoryObj<typeof meta>

export const SplitPane: Story = {}
