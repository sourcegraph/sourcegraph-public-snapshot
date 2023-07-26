<script lang="ts">
    import { mdiCodeBracesBox, mdiFileCodeOutline } from '@mdi/js'

    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import Icon from '$lib/Icon.svelte'
    import LoadingSpinner from '$lib/LoadingSpinner.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import { createPromiseStore } from '$lib/utils'

    import FileDiff from '../../../../-/commit/[...revspec]/FileDiff.svelte'
    import Permalink from '../../../../Permalink.svelte'

    import type { PageData } from './$types'
    import FormatAction from './FormatAction.svelte'
    import WrapLinesAction, { lineWrap } from './WrapLinesAction.svelte'

    type Deferred = PageData['deferred']

    export let data: PageData

    // We use the latest value here because we want to keep showing the old document while loading
    // the new one.
    const { pending: loading, latestValue: blobData, set: setBlob } = createPromiseStore<Deferred['blob']>()
    const { value: highlights, set: setHighlights } = createPromiseStore<Deferred['highlights']>()
    $: setBlob(data.deferred.blob)
    $: setHighlights(data.deferred.highlights)
    $: formatted = !!$blobData?.richHTML
    $: showRaw = $page.url.searchParams.get('view') === 'raw'
</script>

<FileHeader>
    <Icon slot="icon" svgPath={data.deferred.compare ? mdiCodeBracesBox : mdiFileCodeOutline} />
    <svelte:fragment slot="actions">
        {#if data.deferred.compare}
            <span>{data.deferred.compare.revisionToCompare}</span>
        {:else}
            {#if !formatted || showRaw}
                <WrapLinesAction />
            {/if}
            {#if formatted}
                <FormatAction />
            {/if}
            <Permalink resolvedRevision={data.resolvedRevision} />
        {/if}
    </svelte:fragment>
</FileHeader>

<div class="content" class:loading={$loading} class:compare={!!data.deferred.compare}>
    {#if data.deferred.compare}
        {#await data.deferred.compare.diff}
            <LoadingSpinner />
        {:then fileDiff}
            {#if fileDiff}
                <FileDiff {fileDiff} />
            {:else}
                Unable to load iff
            {/if}
        {/await}
    {:else if $blobData}
        {#if $blobData.richHTML && !showRaw}
            <div class="rich">
                {@html $blobData.richHTML}
            </div>
        {:else}
            <!-- TODO: ensure that only the highlights for the currently loaded blob data are used -->
            <CodeMirrorBlob blob={$blobData} highlights={$highlights || ''} wrapLines={$lineWrap} />
        {/if}
    {/if}
</div>

<style lang="scss">
    .content {
        display: flex;
        overflow-x: auto;
        flex: 1;

        &.compare {
            flex-direction: column;
        }
    }
    .loading {
        filter: blur(1px);
    }

    .rich {
        padding: 1rem;
        overflow: auto;
    }
</style>
