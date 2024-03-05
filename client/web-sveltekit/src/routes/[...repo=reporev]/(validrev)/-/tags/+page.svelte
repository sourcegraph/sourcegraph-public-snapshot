<script lang="ts">
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'
    import { Alert } from '$lib/wildcard'

    import type { PageData } from './$types'

    export let data: PageData
</script>

<svelte:head>
    <title>Tags - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <div>
        {#await data.tags}
            <LoadingSpinner />
        {:then connection}
            <!-- TODO: Search input to filter tags by name -->
            <!-- TODO: Pagination -->
            <table>
                <tbody>
                    {#each connection.nodes as node (node.id)}
                        <GitReference ref={node} />
                    {:else}
                        <tr>
                            <td colspan="2">
                                <Alert variant="info">No tags found</Alert>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
            <small class="text-muted">{connection.totalCount} tags total</small>
        {:catch error}
            <Alert variant="danger">{error.message}</Alert>
        {/await}
    </div>
</section>

<style lang="scss">
    table {
        width: 100%;
        border-spacing: 0;
    }

    section {
        overflow: auto;
        margin-top: 2rem;
    }

    div {
        max-width: 54rem;
        margin-left: auto;
        margin-right: auto;
    }
</style>
