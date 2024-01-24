<script lang="ts">
    import Commit from '$lib/Commit.svelte'

    import type { PageData, Snapshot } from './$types'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
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
        restore(snapshot) {
            restoredCommitCount = snapshot.commitCount
            restoredScroller = snapshot.scroller
            scroller.restore(snapshot.scroller)
        },
    }

    /**
     * Fetches more commits when the user scrolls to the bottom of the page.
     */
    function fetchMore() {
        // Only fetch more commits if there are more commits and if we are not already
        // fetching more commits.
        if (commits?.pageInfo.hasNextPage && $commitsQuery && !$commitsQuery.loading) {
            commitsQuery.fetchMore({
                variables: {
                    afterCursor: commits.pageInfo.endCursor,
                },
            })
        }
    }

    /**
     * Restores the previous scroll position when the user refreshes the page. Normally
     * this would bring the user back to the top of the page, but we keep track of how
     * many commits were previously loaded and fetch the missing commits if necessary.
     * It's not ideal because we can only start fetching the remaining data when the
     * component mounts, but it's better than nothing.
     */
    async function restoreCommits(
        commits: CommitsPage_GitCommitConnection | undefined,
        commitCount: number,
        scrollerCapture: ScrollerCapture | undefined
    ) {
        // Fetch more commits to restore the previous scroll position
        if (commits) {
            if (commits.nodes.length < commitCount && !restoring) {
                restoring = true
                await commitsQuery.fetchMore({
                    variables: {
                        afterCursor: commits.pageInfo.endCursor,
                        first: restoredCommitCount - commits.nodes.length,
                    },
                })
                if (scrollerCapture) {
                    scroller.restore(scrollerCapture)
                }
                restoring = false
            }
            restored = true
        }
    }

    let scroller: Scroller
    // The number of commits that were previously loaded. This is only comes into
    // play when the user refreshes the page and thus the Apollo cache is empty.
    let restoredCommitCount: number = 0
    // The previous scroll position. Similiarly this is only used when the user
    // refreshes the page.
    let restoredScroller: ScrollerCapture | undefined
    // Restoring a large number of commits can take a while. This flag is used to
    // show a loading spinner instead of the first page of commits while restoring.
    let restoring = false
    // This flag is used to prevent retrying restoring commits in case of unexpected
    // issues with restoring.
    let restored = false

    $: commitsQuery = data.commitsQuery
    $: commits =
        $commitsQuery?.data.node?.__typename === 'Repository' ? $commitsQuery.data.node.commit?.ancestors : undefined
    $: if (!restored) {
        restoreCommits(commits, restoredCommitCount, restoredScroller)
    }
</script>

<svelte:head>
    <title>Commits - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    <Scroller bind:this={scroller} margin={600} on:more={fetchMore}>
        {#if commits && !restoring}
            <ul>
                {#each commits.nodes as commit (commit.canonicalURL)}
                    <li><Commit {commit} /></li>
                {/each}
            </ul>
        {/if}
        {#if !$commitsQuery || $commitsQuery.loading || restoring}
            <div>
                <LoadingSpinner />
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

    ul {
        list-style: none;
        padding: 1rem;
        max-width: var(--viewport-xl);
        margin: 0 auto;
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
