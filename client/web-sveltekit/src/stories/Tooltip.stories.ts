import type { Meta, StoryObj } from '@storybook/svelte'

import TooltipExample from './TooltipExample.svelte'

const meta: Meta<typeof TooltipExample> = {
    component: TooltipExample,
}

export default meta
type Story = StoryObj<typeof meta>

export const Tooltip: Story = {}
