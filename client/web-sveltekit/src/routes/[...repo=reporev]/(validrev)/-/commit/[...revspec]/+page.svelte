<script lang="ts">
    import Commit from '$lib/Commit.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'
    import FileDiff from './FileDiff.svelte'

    type Deferred = PageData['deferred']

    export let data: PageData

    const { pending: commitPending, value: commit, set: setCommit } = createPromiseStore<Deferred['commit']>()
    $: setCommit(data.deferred.commit)
    const { pending: diffPending, value: diff, set: setDiff } = createPromiseStore<Deferred['diff']>()
    $: setDiff(data.deferred.diff)
    $: pending = $diffPending || $commitPending
</script>

<svelte:head>
    <title>Commit: {$commit?.subject ?? ''} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<section>
    {#if $commit}
        <div class="header">
            <div class="info"><Commit commit={$commit} alwaysExpanded /></div>
            <div>
                <span>Commit:&nbsp;{$commit.abbreviatedOID}</span>
                <span class="parents">
                    {$commit.parents.length} parents:
                    {#each $commit.parents as parent}
                        <a href={parent.canonicalURL}>{parent.abbreviatedOID}</a>{' '}
                    {/each}
                </span>
            </div>
        </div>
        {#if $diff}
            <ul>
                {#each $diff.nodes as node}
                    <li><FileDiff fileDiff={node} /></li>
                {/each}
            </ul>
        {/if}
    {/if}
    {#if pending}
        <LoadingSpinner />
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
