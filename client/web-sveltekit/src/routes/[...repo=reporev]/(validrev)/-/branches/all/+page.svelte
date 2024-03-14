<script lang="ts">
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'

    import type { PageData } from './$types'
    import { Alert } from '$lib/wildcard'

    export let data: PageData
</script>

<svelte:head>
    <title>All branches - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

{#await data.branches}
    <LoadingSpinner />
{:then branches}
    <!-- TODO: Search input to filter branches by name -->
    <!-- TODO: Pagination -->
    <table>
        <tbody>
            {#each branches.nodes as node (node.id)}
                <GitReference ref={node} />
            {/each}
        </tbody>
    </table>
    <small class="text-muted">{branches.totalCount} branches total</small>
{:catch error}
    <Alert variant="danger">
        Unable to fetch branches information:
        {error.message}
    </Alert>
{/await}

<style lang="scss">
    table {
        width: 100%;
        border-spacing: 0;
    }
</style>
