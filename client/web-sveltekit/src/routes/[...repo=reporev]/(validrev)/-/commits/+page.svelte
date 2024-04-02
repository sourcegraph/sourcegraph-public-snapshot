<script lang="ts">
    import Commit from '$lib/Commit.svelte'

    import type { PageData, Snapshot } from './$types'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { get } from 'svelte/store'
    import { navigating } from '$app/stores'
    import { Alert } from '$lib/wildcard'
    import type { CommitsPage_GitCommitConnection } from './page.gql'

    export let data: PageData

    // This tracks the number of commits that have been loaded and the current scroll
    // position, so both can be restored when the user refreshes the page or navigates
    // back to it.
    export const snapshot: Snapshot<{ commitCount: number; scroller: ScrollerCapture }> = {
        capture() {
            return {
                commitCount: commits?.nodes.length ?? 0,
                scroller: scroller.capture(),
            }
        },
        async restore(snapshot) {
            if (snapshot?.commitCount !== undefined && get(navigating)?.type === 'popstate') {
                await commitsQuery?.restore(result => {
                    const count = result.data?.repository?.commit?.ancestors.nodes?.length
                    return !!count && count < snapshot.commitCount
                })
            }
            scroller.restore(snapshot.scroller)
        },
    }

    function fetchMore() {
        commitsQuery?.fetchMore()
    }

    let scroller: Scroller
    let commits: CommitsPage_GitCommitConnection | null = null

    $: commitsQuery = data.commitsQuery
    // We conditionally check for the ancestors field to be able to show
    // previously loaded commits when an error occurs while fetching more commits.
    $: if ($commitsQuery?.data?.repository?.commit?.ancestors) {
        commits = $commitsQuery.data.repository.commit.ancestors
    }
</script>

<svelte:head>
    <title>Commits - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <Scroller bind:this={scroller} margin={600} on:more={fetchMore}>
        {#if !$commitsQuery.restoring && commits}
            <ul>
                {#each commits.nodes as commit (commit.canonicalURL)}
                    <li><Commit {commit} /></li>
                {:else}
                    <li>
                        <Alert variant="info">No commits found</Alert>
                    </li>
                {/each}
            </ul>
        {/if}
        {#if $commitsQuery.fetching || $commitsQuery.restoring}
            <div>
                <LoadingSpinner />
            </div>
        {:else if $commitsQuery.error}
            <div>
                <Alert variant="danger">
                    Unable to fetch commits: {$commitsQuery.error.message}
                </Alert>
            </div>
        {/if}
    </Scroller>
</section>

<style lang="scss">
    section {
        flex: 1;
        min-height: 0;
        overflow: hidden;
    }

    ul,
    div {
        padding: 1rem;
        max-width: var(--viewport-xl);
        margin: 0 auto;
    }

    ul {
        list-style: none;
        --avatar-size: 2.5rem;
    }

    li {
        border-bottom: 1px solid var(--border-color);
        padding: 0.5rem 0;

        &:last-child {
            border: none;
        }
    }

    div {
        &:not(:first-child) {
            border-top: 1px solid var(--border-color);
        }
        padding: 0.5rem 0;
    }
</style>
