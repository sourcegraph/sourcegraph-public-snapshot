<script lang="ts">
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'

    export let data: PageData

    const { pending, value: branches, set } = createPromiseStore<PageData['deferred']['branches']>()
    $: set(data.deferred.branches)
    $: defaultBranch = $branches?.defaultBranch
    $: activeBranches = $branches?.activeBranches
</script>

<svelte:head>
    <title>Branches - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

{#if $pending}
    <LoadingSpinner />
{/if}

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
