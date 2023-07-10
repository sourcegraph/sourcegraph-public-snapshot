<script lang="ts" context="module">
    // Keep in sync with TreeView.stories.ts (can't be exported for some reason)
    interface ExampleData {
        name: string
        children?: ExampleData[]
    }
</script>

<script lang="ts">
    import { writable } from 'svelte/store'

    import { updateTreeState, type TreeProvider, type TreeState, TreeStateUpdate } from '$lib/TreeView'
    import TreeView, { setTreeContext } from '$lib/TreeView.svelte'

    export let treeProvider: TreeProvider<ExampleData>
    export let treeState: TreeState

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

<TreeView {treeProvider} on:select={handleSelect}>
    <svelte:fragment let:entry>
        {entry.name}
    </svelte:fragment>
</TreeView>

<style lang="scss">
    :global(.label:hover),
    :global(.treeitem.selected) > :global(.label) {
        background-color: lightblue;
    }
    :global(.treeitem:focus) > :global(.label) {
        outline: 2px solid green !important;
    }
</style>
