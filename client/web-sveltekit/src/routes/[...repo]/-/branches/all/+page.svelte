<script lang="ts">
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'

    import type { PageData } from './$types'

    export let data: PageData

    $: branches = data.branches
    $: nodes = !$branches.loading && $branches.data ? $branches.data.nodes : null
    $: total = !$branches.loading && $branches.data ? $branches.data.totalCount : null
</script>

{#if $branches.loading}
    <LoadingSpinner />
{:else if nodes}
    <!-- TODO: Search input to filter branches by name -->
    <!-- TODO: Pagination -->
    <table>
        <tbody>
            {#each nodes as node (node.id)}
                <GitReference ref={node} />
            {/each}
        </tbody>
    </table>
    {#if total !== null}
        <small class="text-muted">{total} branches total</small>
    {/if}
{/if}

<style lang="scss">
    table {
        width: 100%;
        border-spacing: 0;
    }
</style>
