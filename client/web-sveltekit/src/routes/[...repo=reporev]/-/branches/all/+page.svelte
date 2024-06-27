<script lang="ts">
    // @sg EnableRollout
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import { pluralize } from '$lib/common'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { Alert, Button, Input } from '$lib/wildcard'
    import type { GitBranchesConnection } from '$testing/graphql-type-mocks'

    import type { PageData, Snapshot } from './$types'

    export let data: PageData

    export const snapshot: Snapshot<{ count: number; scroller: ScrollerCapture }> = {
        capture() {
            return {
                count: branchesConnection?.nodes.length ?? 0,
                scroller: scroller.capture(),
            }
        },
        async restore(snapshot) {
            if (snapshot?.count && get(navigating)?.type === 'popstate') {
                await branchesQuery?.restore(result => {
                    const count = result.data?.repository?.branches?.nodes?.length
                    return !!count && count < snapshot.count
                })
            }
            scroller.restore(snapshot.scroller)
        },
    }

    let scroller: Scroller
    let branchesConnection: GitBranchesConnection | undefined

    $: query = data.query
    $: branchesQuery = data.branchesQuery
    $: branchesConnection = $branchesQuery.data?.repository?.branches ?? branchesConnection
    $: if (branchesQuery) {
        branchesConnection = undefined
    }
</script>

<svelte:head>
    <title>All branches - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <form method="GET">
        <Input type="search" name="query" placeholder="Search branches" value={query} autofocus />
        <Button variant="primary" type="submit">Search</Button>
    </form>
    <Scroller bind:this={scroller} margin={600} on:more={branchesQuery.fetchMore}>
        {#if !$branchesQuery.restoring && branchesConnection}
            <table>
                <tbody>
                    {#each branchesConnection.nodes as tag (tag)}
                        <GitReference ref={tag} />
                    {:else}
                        <tr>
                            <td colspan="2">
                                <Alert variant="info">No tags found</Alert>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        {/if}
        <div>
            {#if $branchesQuery.fetching || $branchesQuery.restoring}
                <LoadingSpinner />
            {:else if $branchesQuery.error}
                <Alert variant="danger">
                    Unable to load branches: {$branchesQuery.error.message}
                </Alert>
            {/if}
        </div>
    </Scroller>
    {#if branchesConnection && branchesConnection.nodes.length > 0}
        <div class="footer">
            {branchesConnection.totalCount}
            {pluralize('branch', branchesConnection.totalCount)} total
            {#if branchesConnection.totalCount > branchesConnection.nodes.length}
                (showing {branchesConnection.nodes.length})
            {/if}
        </div>
    {/if}
</section>

<style lang="scss">
    section {
        display: flex;
        flex-direction: column;
        height: 100%;
        overflow: hidden;

        :global([data-scroller]) {
            display: flex;
            flex-direction: column;
        }

        form,
        div,
        :global([data-scroller]) {
            padding: 1rem;
        }
    }

    form {
        display: flex;
        gap: 1rem;

        :global([data-input-container]) {
            flex: 1;
        }
    }

    form,
    div,
    table {
        align-self: center;
        max-width: var(--viewport-xl);
        width: 100%;
    }

    table {
        border-spacing: 0;
    }

    .footer {
        color: var(--text-muted);
        // Unset `div` width: 100% to allow the footer to be centered
        width: initial;
    }
</style>
