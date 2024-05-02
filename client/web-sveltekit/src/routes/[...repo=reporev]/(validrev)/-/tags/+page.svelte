<script lang="ts">
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import { pluralize } from '$lib/common'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { Alert, Button, Input } from '$lib/wildcard'

    import type { PageData, Snapshot } from './$types'
    import type { GitTagsConnection } from './page.gql'

    export let data: PageData

    export const snapshot: Snapshot<{ count: number; scroller: ScrollerCapture }> = {
        capture() {
            return {
                count: tagsConnection?.nodes.length ?? 0,
                scroller: scroller.capture(),
            }
        },
        async restore(snapshot) {
            if (snapshot?.count && get(navigating)?.type === 'popstate') {
                await tagsQuery?.restore(result => {
                    const count = result.data?.repository?.gitRefs?.nodes?.length
                    return !!count && count < snapshot.count
                })
            }
            scroller.restore(snapshot.scroller)
        },
    }

    let scroller: Scroller
    let tagsConnection: GitTagsConnection | undefined

    $: query = data.query
    $: tagsQuery = data.tagsQuery
    $: tagsConnection = $tagsQuery.data?.repository?.gitRefs ?? tagsConnection
    $: if (tagsQuery) {
        tagsConnection = undefined
    }
</script>

<svelte:head>
    <title>Tags - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <form method="GET">
        <Input type="search" name="query" placeholder="Search tags" value={query} autofocus />
        <Button variant="primary" type="submit">Search</Button>
    </form>
    <Scroller bind:this={scroller} margin={600} on:more={tagsQuery.fetchMore}>
        {#if !$tagsQuery.restoring && tagsConnection}
            <table>
                <tbody>
                    {#each tagsConnection.nodes as tag (tag)}
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
            {#if $tagsQuery.fetching || $tagsQuery.restoring}
                <LoadingSpinner />
            {:else if $tagsQuery.error}
                <Alert variant="danger">
                    Unable to load tags: {$tagsQuery.error.message}
                </Alert>
            {/if}
        </div>
    </Scroller>
    {#if tagsConnection && tagsConnection.nodes.length > 0}
        <div class="footer">
            {tagsConnection.totalCount}
            {pluralize('tag', tagsConnection.totalCount)} total
            {#if tagsConnection.totalCount > tagsConnection.nodes.length}
                (showing {tagsConnection.nodes.length})
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
    :global([data-scroller]) {
        padding: 1rem;
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
