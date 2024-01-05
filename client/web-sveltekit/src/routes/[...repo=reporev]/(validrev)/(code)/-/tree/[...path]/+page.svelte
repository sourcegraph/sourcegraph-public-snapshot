<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import Permalink from '$lib/repo/Permalink.svelte'
    import { createPromiseStore } from '$lib/utils'
    import type { TreeWithCommitInfo } from './page.gql'

    import FileDiff from '../../../../-/commit/[...revspec]/FileDiff.svelte'

    import type { PageData } from './$types'
    import FileTable from '$lib/repo/FileTable.svelte'

    export let data: PageData

    const { value: tree, set: setTree } = createPromiseStore<PageData['deferred']['treeEntries']>()
    const { value: commitInfo, set: setCommitInfo } = createPromiseStore<Promise<TreeWithCommitInfo | null>>()
    const { value: readme, set: setReadme } = createPromiseStore<PageData['deferred']['readme']>()

    $: setTree(data.deferred.treeEntries)
    $: setCommitInfo(data.deferred.commitInfo)
    $: setReadme(data.deferred.readme)
    $: entries = $tree?.entries ?? []
    $: entriesWithCommitInfo = $commitInfo?.entries ?? []
</script>

<svelte:head>
    <title>{data.filePath} - {data.displayRepoName} - Sourcegraph</title>
</svelte:head>

<FileHeader>
    <Icon slot="icon" svgPath={mdiFolderOutline} />
    <svelte:fragment slot="actions">
        <Permalink resolvedRevision={data.resolvedRevision} />
    </svelte:fragment>
</FileHeader>

<div class="content">
    {#if data.deferred.compare}
        {#await data.deferred.compare.diff}
            <LoadingSpinner />
        {:then nodes}
            {#each nodes as fileDiff}
                <FileDiff {fileDiff} expanded={false} />
            {/each}
        {/await}
    {:else}
        <FileTable revision={data.revision ?? ''} {entries} commitInfo={entriesWithCommitInfo} />
    {/if}
    {#if $readme}
        <h4 class="header">
            <Icon svgPath={mdiFileDocumentOutline} />
            &nbsp;
            {$readme.name}
        </h4>
        <div class="readme">
            {#if $readme.richHTML}
                {@html $readme.richHTML}
            {:else if $readme.content}
                <pre>{$readme.content}</pre>
            {/if}
        </div>
    {/if}
</div>

<style lang="scss">
    .content {
        flex: 1;
    }

    .header {
        background-color: var(--body-bg);
        position: sticky;
        top: 2.8rem;
        padding: 0.5rem;
        border-bottom: 1px solid var(--border-color);
        border-top: 1px solid var(--border-color);
        margin: 0;
    }

    .readme {
        padding: 1rem;
        flex: 1;

        pre {
            white-space: pre-wrap;
        }
    }
</style>
