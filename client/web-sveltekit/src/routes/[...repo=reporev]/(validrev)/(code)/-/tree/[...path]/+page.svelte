<script lang="ts">
    import { mdiFileDocumentOutline, mdiFolderOutline } from '@mdi/js'

    import { isErrorLike } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import Permalink from '$lib/repo/Permalink.svelte'
    import { createPromiseStore } from '$lib/utils'

    import FileDiff from '../../../../-/commit/[...revspec]/FileDiff.svelte'

    import type { PageData } from './$types'

    export let data: PageData

    const { value: treeOrError, set } = createPromiseStore<PageData['deferred']['treeEntries']>()
    $: set(data.deferred.treeEntries)
    $: entries = $treeOrError && !isErrorLike($treeOrError) ? $treeOrError.entries : []
</script>

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
        {:then result}
            {#each result.nodes as fileDiff}
                <FileDiff {fileDiff} expanded={false} />
            {/each}
        {/await}
    {:else}
        <ul>
            {#each entries as entry}
                <li>
                    <a href={entry.url}>
                        <Icon svgPath={entry.isDirectory ? mdiFolderOutline : mdiFileDocumentOutline} inline />
                        {entry.name}
                    </a>
                </li>
            {/each}
        </ul>
    {/if}
</div>

<style lang="scss">
    .content {
        padding: 1rem;
        flex: 1;
    }

    ul {
        list-style: none;
        padding: 0;
        margin: 0;
    }
</style>
