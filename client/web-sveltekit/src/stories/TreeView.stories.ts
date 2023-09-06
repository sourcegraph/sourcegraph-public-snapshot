import type { Meta, StoryObj } from '@storybook/svelte'

import { createEmptySingleSelectTreeState, type TreeProvider } from '$lib/TreeView'

import TreeViewExample from './TreeView.example.svelte'

// Keep in sync with TreeView.example.svelte (can't be imported for some reason)
interface ExampleData {
    name: string
    children?: ExampleData[]
}

class ExampleProvider implements TreeProvider<ExampleData> {
    constructor(private nodes: ExampleData[], private parentPath: string = '') {}
    isSelectable(_entry: ExampleData): boolean {
        return true
    }
    isExpandable(entry: ExampleData): boolean {
        return !!entry.children
    }
    getNodeID(entry: ExampleData): string {
        return this.parentPath + entry.name
    }
    getEntries(): ExampleData[] {
        return this.nodes
    }
    fetchChildren(entry: ExampleData): Promise<TreeProvider<ExampleData>> {
        return Promise.resolve(new ExampleProvider(entry.children ?? [], `${this.parentPath}${entry.name}/`))
    }
}

type TreeConfig = [number, ...(TreeConfig | undefined)[]]

function makeExampleData(config: TreeConfig, level = 0): ExampleData[] {
    const [n, ...children] = config
    return Array.from({ length: n }, (_, i) => ({
        name: `level${level}-${i + 1}`,
        children: children[i] ? makeExampleData(children[i]!, level + 1) : undefined,
    }))
}

const meta = {
    component: TreeViewExample,
} satisfies Meta<TreeViewExample>

export default meta
type Story = StoryObj<typeof meta>

export const Simple: Story = {
    render: args => ({
        Component: TreeViewExample,
        props: args,
    }),
    args: {
        treeProvider: new ExampleProvider(makeExampleData([3, [2], [3]])),
        treeState: createEmptySingleSelectTreeState(),
    },
}

export const DeeplyNested: Story = {
    render: args => ({
        Component: TreeViewExample,
        props: args,
    }),
    args: {
        treeProvider: new ExampleProvider(
            makeExampleData([5, [3, [2, [2, [3]]], [1], [2, [1, [2]]]], , [3, [2, [2], [3]], [1], [2, [1, [2]]]], [3]])
        ),
        treeState: createEmptySingleSelectTreeState(),
    },
}
