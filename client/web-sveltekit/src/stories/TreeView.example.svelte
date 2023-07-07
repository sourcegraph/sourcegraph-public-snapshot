<script lang="ts">
    import { updateNodeState, type TreeProvider, type TreeState } from '$lib/TreeView'
    import TreeView from '$lib/TreeView.svelte'

    export let treeProvider: TreeProvider<string>
    export let treeState: TreeState

    function handleSelect({ detail }: { detail: Element }) {
        treeState = updateNodeState(treeState, (detail as HTMLElement).dataset.nodeId ?? '', { selected: true })
    }
</script>

<TreeView {treeProvider} bind:treeState isRoot on:select={handleSelect}>
    <svelte:fragment let:entry>
        {entry}
    </svelte:fragment>
</TreeView>
