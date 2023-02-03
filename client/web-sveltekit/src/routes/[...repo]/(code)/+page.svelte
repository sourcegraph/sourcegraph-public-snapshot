<script lang="ts">
    import { mdiFolderOutline, mdiFileDocumentOutline } from '@mdi/js'

    import { isErrorLike } from '$lib/common'
    import Commit from '$lib/Commit.svelte'
    import Icon from '$lib/Icon.svelte'

    import type { PageData } from './$types'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'

    export let data: PageData

    $: treeOrError = data.treeEntries
    $: commits = data.commits
</script>

<div class="content">
    {#if !isErrorLike(data.resolvedRevision)}
        <h3>Description</h3>
        <p>
            {data.resolvedRevision.repo.description}
        </p>
    {/if}

    {#if !$treeOrError.loading && $treeOrError.data && !isErrorLike($treeOrError.data)}
        <h3>Files and directories</h3>
        <ul class="files">
            {#each $treeOrError.data.entries as entry}
                <li>
                    <a
                        data-sveltekit-preload-data={entry.isDirectory ? 'hover' : 'tap'}
                        data-sveltekit-preload-code="hover"
                        href={entry.url}
                        ><Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                        {entry.name}</a
                    >
                </li>
            {/each}
        </ul>
    {/if}

    <h3 class="mt-3">Changes</h3>
    <ul class="commits">
        {#if $commits.loading}
            <LoadingSpinner />
        {:else if $commits.data}
            {#each $commits.data as commit (commit.url)}
                <li><Commit {commit} /></li>
            {/each}
        {/if}
    </ul>
</div>

<style lang="scss">
    .content {
        padding: 1rem;
        overflow: auto;
    }

    ul.commits {
        padding: 0;
        margin: 0;
        list-style: none;

        li {
            border-bottom: 1px solid var(--border-color);
            padding: 0.5rem 0;

            &:last-child {
                border: none;
            }
        }
    }

    ul.files {
        padding: 0;
        margin: 0;
        list-style: none;
        columns: 3;
    }
</style>
