import type { Meta, StoryObj } from '@storybook/svelte'
import type { ComponentProps } from 'svelte'

import type { TreeProvider } from '$lib/TreeView'
import TreeView from '$lib/TreeView.svelte'

import TreeViewExample from './TreeView.example.svelte'

const meta: Meta<ComponentProps<TreeView>> = {
    component: TreeView,
    args: {
        treeProvider: new (class implements TreeProvider<string> {
            isExpandable(entry: string): boolean {
                return false
            }
            getKey(entry: string): string {
                return entry
            }
            getEntries(): string[] {
                return ['foo', 'bar', 'baz']
            }
            fetchChildren(entry: string): Promise<TreeProvider<string>> {
                return Promise.resolve(new this.constructor())
            }
        })(),
        treeState: { focused: 'foo', nodes: {} },
    },
}

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
    render: args => ({
        Component: TreeViewExample,
        props: args,
    }),
}
