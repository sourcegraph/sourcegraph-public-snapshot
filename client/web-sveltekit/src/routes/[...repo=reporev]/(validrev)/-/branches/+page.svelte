<script lang="ts">
    // @sg EnableRollout
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'

    import type { PageData } from './$types'
    import { Alert } from '$lib/wildcard'

    export let data: PageData
</script>

<svelte:head>
    <title>Branches - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <div>
        {#await data.overview}
            <LoadingSpinner />
        {:then result}
            {@const activeBranches = result.branches.nodes.filter(branch => branch.id !== result.defaultBranch?.id)}

            {#if result.defaultBranch}
                <table class="mb-3">
                    <thead><tr><th colspan="3">Default branch</th></tr></thead>
                    <tbody>
                        <GitReference ref={result.defaultBranch} />
                    </tbody>
                </table>
            {/if}

            {#if activeBranches.length > 0}
                <table>
                    <thead><tr><th colspan="3">Active branches</th></tr></thead>
                    <tbody>
                        {#each activeBranches as branch (branch.id)}
                            <GitReference ref={branch} />
                        {/each}
                    </tbody>
                </table>
            {/if}
        {:catch error}
            <Alert variant="danger">
                Unable to fetch branches:
                {error.message}
            </Alert>
        {/await}
    </div>
</section>

<style lang="scss">
    section {
        overflow: auto;
    }

    div {
        max-width: var(--viewport-xl);
        width: 100%;
        margin: 0 auto;

        padding: 1rem;
    }

    table {
        width: 100%;
        border: 1px solid var(--border-color-2);
        background-color: var(--color-bg-1);
        border-radius: var(--border-radius);
        border-spacing: 0;
    }

    thead th {
        font-weight: normal;
        padding: 0.5rem;
        background-color: var(--color-bg-2);
    }
</style>
