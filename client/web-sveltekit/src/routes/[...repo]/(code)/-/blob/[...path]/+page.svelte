<script lang="ts">
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import type { PageData } from './$types'
    import { page } from '$app/stores'
    import WrapLinesAction, { lineWrap } from './WrapLinesAction.svelte'
    import FormatAction from './FormatAction.svelte'
    import type { BlobFileFields } from '$lib/graphql-operations'
    import HeaderAction from '$lib/repo/HeaderAction.svelte'

    export let data: PageData

    $: blob = data.blob
    $: highlights = data.highlights
    $: loading = $blob.loading
    let blobData: BlobFileFields
    $: if ($blob && !$blob.loading && $blob.data) {
        blobData = $blob.data
    }
    $: formatted = !!blobData?.richHTML
    $: showRaw = $page.url.searchParams.get('view') === 'raw'
</script>

{#if !formatted || showRaw}
    <HeaderAction key="wrap-lines" priority={0} component={WrapLinesAction} />
{/if}
{#if formatted}
    <HeaderAction key="format" priority={-1} component={FormatAction} />
{/if}
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
        overflow: hidden;
        display: flex;
    }
    .loading {
        filter: blur(1px);
    }

    .rich {
        padding: 1rem;
        overflow: auto;
    }
</style>
