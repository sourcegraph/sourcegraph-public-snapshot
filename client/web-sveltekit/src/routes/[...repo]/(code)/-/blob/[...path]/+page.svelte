<script lang="ts">
    import { mdiFileCodeOutline } from '@mdi/js'

    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import type { BlobFileFields } from '$lib/graphql-operations'
    import Icon from '$lib/Icon.svelte'
    import FileHeader from '$lib/repo/FileHeader.svelte'
    import { asStore } from '$lib/utils'

    import type { PageData } from './$types'
    import FormatAction from './FormatAction.svelte'
    import WrapLinesAction, { lineWrap } from './WrapLinesAction.svelte'

    export let data: PageData

    $: blob = asStore(data.blob.deferred)
    $: highlights = asStore(data.highlights.deferred)
    $: loading = $blob.loading
    let blobData: BlobFileFields
    $: if (!$blob.loading && $blob.data) {
        blobData = $blob.data
    }
    $: formatted = !!blobData?.richHTML
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

<div class="content" class:loading>
    {#if blobData}
        {#if blobData.richHTML && !showRaw}
            <div class="rich">
                {@html blobData.richHTML}
            </div>
        {:else}
            <CodeMirrorBlob
                blob={blobData}
                highlights={($highlights && !$highlights.loading && $highlights.data) || ''}
                wrapLines={$lineWrap}
            />
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
