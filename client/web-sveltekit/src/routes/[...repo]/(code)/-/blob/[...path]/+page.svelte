<script lang="ts">
    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import type { BlobFileFields } from '$lib/graphql-operations'
    import HeaderAction from '$lib/repo/HeaderAction.svelte'
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
