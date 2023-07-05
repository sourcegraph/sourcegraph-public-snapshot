<script lang="ts">
    import { page } from '$app/stores'
    import CodeMirrorBlob from '$lib/CodeMirrorBlob.svelte'
    import Icon from '$lib/Icon.svelte'
    import type { BlobFileFields } from '$lib/graphql-operations'
    import FileHeader from '$lib/repo/ui/FileHeader.svelte'
    import { mdiClose, mdiFileOutline } from '@mdi/js'

    import type { PageData } from './$types'
    import FormatAction from './FormatAction.svelte'
    import WrapLinesAction, { lineWrap } from './WrapLinesAction.svelte'
    import { asStore } from '$lib/utils'
    import { readable } from 'svelte/store'
    import FileDiff from '../../../../-/commit/[...revspec]/FileDiff.svelte'
    import Shimmer from '$lib/Shimmer.svelte'
    import Button from '$lib/wildcard/Button.svelte'

    export let data: PageData

    $: diffMode = !!data.deferred.diff
    $: blob = data.deferred.blob ? asStore(data.deferred.blob) : readable(null)
    $: diff = diffMode ? asStore(data.deferred.diff) : readable(null)
    $: highlights = asStore(data.deferred.highlights)
    $: loading = $blob?.loading
    let blobData: BlobFileFields|null = null
    $: if (!$blob?.loading && $blob?.data) {
        blobData = $blob?.data
    }
    $: formatted = !!blobData?.richHTML
    $: showRaw = $page.url.searchParams.get('view') === 'raw'

    function getURLWithoutCompare(): string {
        const url = new URL($page.url)
        url.searchParams.delete('rev')
        return url.toString()
    }
</script>

<div class="header">
    <FileHeader commit={data.deferred.history}>
        <Icon slot="icon" svgPath={mdiFileOutline} />
        <svelte:fragment slot="actions">
            {#if !diffMode}
                {#if !formatted || showRaw}
                    <WrapLinesAction />
                {/if}
                {#if formatted}
                    <FormatAction />
                {/if}
            {/if}
            {#if diffMode}
                <a href={getURLWithoutCompare()}>
                    <Icon svgPath={mdiClose} inline />
                </a>
            {/if}
        </svelte:fragment>
    </FileHeader>
</div>
<div class="content" class:loading>
    {#if diffMode}
        {#if $diff && $diff.loading}
            <Shimmer />
        {/if}
        {#if $diff && !$diff.loading}
            <div style="overflow: auto">
                <FileDiff fileDiff={$diff.data.nodes[0]}/>
            </div>
        {/if}
    {:else if blobData}
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
    .header {
        background-color: var(--color-bg-1);
        border: 1px solid var(--border-color);
        padding: 0.5rem;
        border-top-right-radius: var(--border-radius);
        border-top-left-radius: var(--border-radius);
    }
    .content {
        overflow: hidden;
        display: flex;
        flex: 1;
        flex-direction: column;
        background-color: var(--code-bg);
        border-left: 1px solid var(--border-color);
        border-right: 1px solid var(--border-color);
    }
    .loading {
        filter: blur(1px);
    }

    .rich {
        padding: 1rem;
        overflow: auto;
    }
</style>
