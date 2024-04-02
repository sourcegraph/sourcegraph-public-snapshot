<script lang="ts">
    import Commit from '$lib/Commit.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'

    import type { PageData, Snapshot } from './$types'
    import FileDiff from '$lib/repo/FileDiff.svelte'
    import Scroller, { type Capture as ScrollerCapture } from '$lib/Scroller.svelte'
    import { get } from 'svelte/store'
    import { navigating } from '$app/stores'
    import type { CommitPage_DiffConnection } from './page.gql'
    import { Alert } from '$lib/wildcard'

    interface Capture {
        scroll: ScrollerCapture
        diffCount: number
        expandedDiffs: Array<[number, boolean]>
    }

    export let data: PageData

    export const snapshot: Snapshot<Capture> = {
        capture: () => ({
            scroll: scroller.capture(),
            diffCount: diffs?.nodes.length ?? 0,
            expandedDiffs: Array.from(expandedDiffs.entries()),
        }),
        restore: async capture => {
            expandedDiffs = new Map(capture.expandedDiffs)
            if (capture?.diffCount !== undefined && get(navigating)?.type === 'popstate') {
                await data.diff?.restore(result => {
                    const count = result.data?.repository?.comparison.fileDiffs.nodes.length
                    return !!count && count < capture.diffCount
                })
            }
            scroller.restore(capture.scroll)
        },
    }

    let scroller: Scroller
    let expandedDiffs = new Map<number, boolean>()
    let diffs: CommitPage_DiffConnection | null = null

    $: diffQuery = data.diff
    // We conditionally check for the ancestors field to be able to show
    // previously loaded commits when an error occurs while fetching more commits.
    $: if ($diffQuery?.data?.repository) {
        diffs = $diffQuery.data.repository.comparison.fileDiffs
    }
</script>

<svelte:head>
    <title>Commit: {data.commit?.subject ?? ''} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    {#if data.commit}
        <Scroller bind:this={scroller} margin={600} on:more={data.diff?.fetchMore}>
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
            {#if !$diffQuery?.restoring && diffs}
                <ul>
                    {#each diffs.nodes as node, index}
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
            {#if $diffQuery?.fetching || $diffQuery?.restoring}
                <LoadingSpinner />
            {:else if $diffQuery?.error}
                <div class="m-4">
                    <Alert variant="danger">
                        Unable to fetch file diffs: {$diffQuery.error.message}
                    </Alert>
                </div>
            {/if}
        </Scroller>
    {/if}
</section>

<style lang="scss">
    section {
        overflow: auto;
    }

    .header {
        display: flex;
        padding: 1rem;
        border-bottom: 1px solid var(--border-color);
    }

    .parents {
        white-space: nowrap;
    }
    .info {
        flex: 1;
    }

    ul {
        list-style: none;
        padding: 1rem;

        li {
            margin-bottom: 1rem;
        }
    }
</style>
