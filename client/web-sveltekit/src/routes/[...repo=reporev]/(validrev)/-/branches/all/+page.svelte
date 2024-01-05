<script lang="ts">
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    const { pending, value: branches, set } = createPromiseStore<PageData['deferred']['branches']>()
    $: set(data.deferred.branches)
    $: nodes = $branches?.nodes
    $: total = $branches?.totalCount
</script>

<svelte:head>
    <title>All branches - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

{#if $pending}
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
