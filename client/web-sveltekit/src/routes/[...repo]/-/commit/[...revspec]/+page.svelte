<script lang="ts">
    import Commit from '$lib/Commit.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import { asStore } from '$lib/utils'

    import type { PageData } from './$types'
    import FileDiff from './FileDiff.svelte'

    export let data: PageData

    $: commit = asStore(data.commit.deferred)
    $: diff = asStore(data.diff.deferred)
</script>

<section>
    {#if !$commit.loading && $commit.data}
        <div class="header">
            <div class="info"><Commit commit={$commit.data} alwaysExpanded /></div>
            <div>
                <span>Commit:&nbsp;{$commit.data.abbreviatedOID}</span>
                <span class="parents">
                    {$commit.data.parents.length} parents:
                    {#each $commit.data.parents as parent}
                        <a href={parent.url}>{parent.abbreviatedOID}</a>{' '}
                    {/each}
                </span>
            </div>
        </div>
        {#if !$diff.loading && $diff.data}
            <ul>
                {#each $diff.data.nodes as node}
                    <li><FileDiff fileDiff={node} /></li>
                {/each}
            </ul>
        {/if}
    {/if}
    {#if $commit.loading || $diff.loading}
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
