<script lang="ts">
    import { mdiFileCodeOutline } from '@mdi/js'

    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import Icon from '$lib/Icon.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import { createPromiseStore } from '$lib/utils'

    import type { PageData } from './$types'
    import FormatAction from './FormatAction.svelte'
    import WrapLinesAction, { lineWrap } from './WrapLinesAction.svelte'

    export let data: PageData

    // We use the latest value here because we want to keep showing the old document while loading
    // the new one.
    const { pending: loading, latestValue: blobData, set: setBlob } = createPromiseStore<typeof data.blob.deferred>()
    const { value: highlights, set: setHighlights } = createPromiseStore<typeof data.highlights.deferred>()
    $: setBlob(data.blob.deferred)
    $: setHighlights(data.highlights.deferred)
    $: formatted = !!$blobData?.richHTML
    $: showRaw = $page.url.searchParams.get('view') === 'raw'
</script>

<FileHeader>
    <Icon slot="icon" svgPath={mdiFileCodeOutline} />
    <svelte:fragment slot="actions">
        {#if !formatted || showRaw}
            <WrapLinesAction />
        {/if}
        {#if formatted}
            <FormatAction />
        {/if}
    </svelte:fragment>
</FileHeader>

<div class="content" class:loading={$loading}>
    {#if $blobData}
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
    }
    .loading {
        filter: blur(1px);
    }

    .rich {
        padding: 1rem;
        overflow: auto;
    }
</style>
