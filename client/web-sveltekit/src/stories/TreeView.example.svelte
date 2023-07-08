<script lang="ts" context="module">
    // Keep in sync with TreeView.stories.ts (can't be exported for some reason)
    interface ExampleData {
        name: string
        children?: ExampleData[]
    }
</script>

<script lang="ts">
    import { updateNodeState, type TreeProvider, type TreeState } from '$lib/TreeView'
    import TreeView from '$lib/TreeView.svelte'

    export let treeProvider: TreeProvider<ExampleData>
    export let treeState: TreeState

    let selected: string

    function handleSelect({ detail: node }: { detail: HTMLElement }) {
        if (selected) {
            treeState = {...treeState, nodes: updateNodeState(treeState, selected, { selected: false })}
        }
        const nodeId = node.dataset.nodeId
        if (nodeId) {
            treeState = {
                ...treeState,
                focused: nodeId,
                nodes: updateNodeState(treeState, node.dataset.nodeId ?? '', { selected: true, expanded: true })
            }
            selected = nodeId
            node.focus()
        }
    }
</script>

<TreeView {treeProvider} bind:treeState isRoot on:select={handleSelect}>
    <svelte:fragment let:entry>
        {entry.name}
    </svelte:fragment>
</TreeView>

<style lang="scss">
    :global(.label:hover), :global(.treeitem.selected) > :global(.label) {
        background-color: lightblue;
    }
    :global(.treeitem:focus) > :global(.label) {
        outline: 2px solid green !important;
    }
</style>
