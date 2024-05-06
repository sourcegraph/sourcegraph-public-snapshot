<script lang="ts" context="module">
    import { Story, Template } from '@storybook/addon-svelte-csf'

    import {
        createEmptySingleSelectTreeState,
        updateTreeState,
        type TreeProvider,
        type TreeState,
        TreeStateUpdate,
    } from '$lib/TreeView'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'

    export const meta = {
        component: TreeView,
    }

    // Keep in sync with TreeView.stories.ts (can't be exported for some reason)
    interface ExampleData {
        name: string
        children?: ExampleData[]
    }

    class ExampleProvider implements TreeProvider<ExampleData> {
        constructor(private nodes: ExampleData[], private parentPath: string = '') {}
        public isSelectable(_entry: ExampleData): boolean {
            return true
        }
        public isExpandable(entry: ExampleData): boolean {
            return !!entry.children
        }
        public getNodeID(entry: ExampleData): string {
            return this.parentPath + entry.name
        }
        public getEntries(): ExampleData[] {
            return this.nodes
        }
        public fetchChildren(entry: ExampleData): Promise<TreeProvider<ExampleData>> {
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
</script>

<script lang="ts">
    import { writable } from 'svelte/store'

    const treeState: TreeState = createEmptySingleSelectTreeState()
    const treeStateStore = writable(treeState)
    setTreeContext(treeStateStore)

    $: $treeStateStore = treeState

    function handleSelect({ detail: node }: { detail: HTMLElement }) {
        const nodeId = node.dataset.nodeId
        if (nodeId) {
            $treeStateStore = updateTreeState(
                $treeStateStore,
                node.dataset.nodeId ?? '',
                TreeStateUpdate.SELECT | TreeStateUpdate.EXPAND
            )
            node.focus()
        }
    }
</script>

<Template let:args>
    <TreeView treeProvider={new ExampleProvider(makeExampleData(args.data))} on:select={handleSelect}>
        <svelte:fragment let:entry>
            {entry.name}
        </svelte:fragment>
    </TreeView>
</Template>

<Story
    name="Simple"
    args={{
        data: [3, [2], [3]],
    }}
/>

<Story
    name="DeeplyNested"
    args={{
        data: [5, [3, [2, [2, [3]]], [1], [2, [1, [2]]]], , [3, [2, [2], [3]], [1], [2, [1, [2]]]], [3]],
    }}
/>

<style lang="scss">
    :global(.label:hover),
    :global([data-treeitem][aria-selected]) > :global(.label) {
        background-color: lightblue;
    }
    :global([data-treeitem]:focus) > :global(.label) {
        outline: 2px solid green !important;
    }
</style>
