<script lang="ts">
    import { get } from 'svelte/store'

    import { navigating } from '$app/stores'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import GitReference from '$lib/repo/GitReference.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { Alert } from '$lib/wildcard'

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
    let tagsConnection: GitTagsConnection

    $: tagsQuery = data.tagsQuery
    $: tagsConnection = $tagsQuery.data?.repository?.gitRefs ?? tagsConnection
</script>

<svelte:head>
    <title>Tags - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    {#if !$tagsQuery.restoring && tagsConnection}
        <Scroller bind:this={scroller} margin={600} on:more={tagsQuery.fetchMore}>
            <!-- TODO: Search input to filter tags by name -->
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
        <div class="footer">
            {tagsConnection.totalCount} tags total (showing {tagsConnection.nodes.length})
        </div>
    {/if}
</section>

<style lang="scss">
    div,
    table {
        max-width: var(--viewport-xl);
        margin: 0 auto;
    }

    table {
        width: 100%;
        border-spacing: 0;
    }

    section {
        display: flex;
        flex-direction: column;
        margin-top: 2rem;
        height: 100%;
        overflow: hidden;
    }

    .footer {
        color: var(--text-muted);
        padding: 1rem;
    }
</style>
