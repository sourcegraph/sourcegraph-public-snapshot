<script lang="ts">
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'

    import type { PageData } from './$types'

    export let data: PageData

    $: tags = data.tags
    $: nodes = !$tags.loading && $tags.data ? $tags.data.nodes : null
    $: total = !$tags.loading && $tags.data ? $tags.data.totalCount : null
</script>

<section>
    <div>
        {#if $tags.loading}
            <LoadingSpinner />
        {:else if nodes}
            <!-- TODO: Search input to filter tags by name -->
            <!-- TODO: Pagination -->
            <table>
                <tbody>
                    {#each nodes as node (node.id)}
                        <GitReference ref={node} />
                    {/each}
                </tbody>
            </table>
            {#if total !== null}
                <small class="text-muted">{total} tags total</small>
            {/if}
        {/if}
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
