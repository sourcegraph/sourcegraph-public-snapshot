<script lang="ts">
    import Commit from '$lib/Commit.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'

    import type { PageData, Snapshot } from './$types'
    import FileDiff from '$lib/repo/FileDiff.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'

    export let data: PageData

    export const snapshot: Snapshot<{ scroll: ScrollerCapture; expandedDiffs: Array<[number, boolean]> }> = {
        capture: () => ({
            scroll: scroller.capture(),
            expandedDiffs: Array.from(expandedDiffs.entries()),
        }),
        restore: capture => {
            scroller.restore(capture.scroll)
            expandedDiffs = new Map(capture.expandedDiffs)
        },
    }

    const diff = data.diff
    let scroller: Scroller
    let loading = true
    let expandedDiffs = new Map<number, boolean>()

    $: fileDiffConnection = $diff?.data.repository?.comparison.fileDiffs ?? null
    $: if ($diff?.data.repository) {
        loading = false
    }

    function fetchMore() {
        if (fileDiffConnection?.pageInfo.hasNextPage) {
            loading = true
            diff?.fetchMore({
                variables: {
                    after: fileDiffConnection.pageInfo.endCursor,
                },
            })
        }
    }
</script>

<svelte:head>
    <title>Commit: {data.commit?.subject ?? ''} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    {#if data.commit}
        <Scroller bind:this={scroller} margin={600} on:more={fetchMore}>
            <div class="header">
                <div class="info"><Commit commit={data.commit} alwaysExpanded /></div>
                <div>
                    <span>Commit:&nbsp;{data.commit.abbreviatedOID}</span>
                    <span class="parents">
                        {data.commit.parents.length} parents:
                        {#each data.commit.parents as parent}
                            <a href={parent.canonicalURL}>{parent.abbreviatedOID}</a>{' '}
                        {/each}
                    </span>
                </div>
            </div>
            {#if fileDiffConnection}
                <ul>
                    {#each fileDiffConnection.nodes as node, index}
                        <li>
                            <FileDiff
                                fileDiff={node}
                                expanded={expandedDiffs.get(index)}
                                on:toggle={event => expandedDiffs.set(index, event.detail.expanded)}
                            />
                        </li>
                    {/each}
                </ul>
            {/if}
            {#if loading}
                <LoadingSpinner />
            {/if}
        </Scroller>
    {/if}
</section>

<style lang="scss">
    section {
        padding: 1rem;
        overflow: auto;
    }

    .header {
        display: flex;
    }

    .parents {
        white-space: nowrap;
    }
    .info {
        flex: 1;
    }

    ul {
        list-style: none;

        li {
            margin-bottom: 1rem;
        }
    }
</style>
