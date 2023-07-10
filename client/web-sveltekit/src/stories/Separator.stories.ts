import type { Meta, StoryObj } from '@storybook/svelte'

import Separator from '$lib/Separator.svelte'

import SeparatorExample from './SeparatorExample.svelte'

const meta: Meta<typeof SeparatorExample> = {
    component: Separator,
}

export default meta
type Story = StoryObj<typeof meta>

export const SplitPane: Story = {
    render: () => ({
        Component: SeparatorExample,
    }),
}
