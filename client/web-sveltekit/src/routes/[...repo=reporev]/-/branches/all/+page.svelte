<script lang="ts">
    // @sg EnableRollout
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import { pluralize } from '$lib/common'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReferencesTable from '$lib/repo/GitReferencesTable.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { Alert, Button, Input } from '$lib/wildcard'

    import type { PageData, Snapshot } from './$types'
    import { GitRefType } from '$lib/graphql-types'

    export let data: PageData

    export const snapshot: Snapshot<{
        branches: ReturnType<typeof data.branchesQuery.capture>
        scroller: ScrollerCapture
    }> = {
        capture() {
            return {
                branches: data.branchesQuery.capture(),
                scroller: scroller.capture(),
            }
        },
        async restore(snapshot) {
            if (get(navigating)?.type === 'popstate') {
                await data.branchesQuery?.restore(snapshot.branches)
            }
            scroller.restore(snapshot.scroller)
        },
    }

    let scroller: Scroller

    $: query = data.query
    $: branchesQuery = data.branchesQuery
    $: branches = $branchesQuery.data
</script>

<svelte:head>
    <title>All branches - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<form method="GET">
    <Input type="search" name="query" placeholder="Search branches" value={query} autofocus />
    <Button variant="primary" type="submit">Search</Button>
</form>
<Scroller bind:this={scroller} margin={600} on:more={branchesQuery.fetchMore}>
    <div class="main">
        {#if branches && branches.nodes.length > 0}
            <GitReferencesTable references={branches.nodes} referenceType={GitRefType.GIT_BRANCH} />
        {/if}
        <div>
            {#if $branchesQuery.fetching}
                <LoadingSpinner />
            {:else if $branchesQuery.error}
                <Alert variant="danger">
                    Unable to load branches: {$branchesQuery.error.message}
                </Alert>
            {:else if !branches || branches.nodes.length === 0}
                <Alert variant="info">No branches found</Alert>
            {/if}
        </div>
    </div>
</Scroller>
{#if branches && branches.nodes.length > 0}
    <div class="footer">
        {branches.totalCount}
        {pluralize('branch', branches.totalCount, 'branches')} total
        {#if branches.totalCount > branches.nodes.length}
            (showing {branches.nodes.length})
        {/if}
    </div>
{/if}

<style lang="scss">
    form {
        display: flex;
        gap: 1rem;
        margin-bottom: 1rem;

        :global([data-input-container]) {
            flex: 1;
        }
    }

    form,
    .main {
        margin: 0 auto;
        max-width: var(--viewport-xl);
        width: 100%;
        padding: 0 1rem;
    }

    .footer {
        color: var(--text-muted);
        // Unset `div` width: 100% to allow the footer to be centered
        width: initial;
    }

    @media (--mobile) {
        .main {
            padding: 0;
        }
    }
</style>
