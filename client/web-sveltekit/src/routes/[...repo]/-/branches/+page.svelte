<script lang="ts">
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'
    import type { PageData } from './$types'

    export let data: PageData

    $: branchesData = data.branches
    $: defaultBranch = !$branchesData.loading && $branchesData.data ? $branchesData.data.defaultBranch : null
    $: activeBranches = !$branchesData.loading && $branchesData.data ? $branchesData.data.activeBranches : null
</script>

{#if $branchesData.loading}
    <LoadingSpinner />
{:else if $branchesData.data}
    {#if defaultBranch}
        <table class="mb-3">
            <thead><tr><th colspan="3">Default branch</th></tr></thead>
            <tbody>
                <GitReference ref={defaultBranch} />
            </tbody>
        </table>
    {/if}

    {#if activeBranches && activeBranches.length > 0}
        <table>
            <thead><tr><th colspan="3">Active branches</th></tr></thead>
            <tbody>
                {#each activeBranches as branch (branch.id)}
                    <GitReference ref={branch} />
                {/each}
            </tbody>
        </table>
    {/if}
{/if}

<style lang="scss">
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
