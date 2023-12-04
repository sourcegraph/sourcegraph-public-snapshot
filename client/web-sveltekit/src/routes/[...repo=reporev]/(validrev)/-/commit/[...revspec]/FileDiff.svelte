<script lang="ts">
    import { dirname } from 'path'

    import { mdiChevronRight, mdiChevronDown } from '@mdi/js'

    import { numberWithCommas } from '$lib/common'
    import type { FileDiffFields } from '$lib/graphql-operations'
    import Icon from '$lib/Icon.svelte'

    import DiffSquares from './DiffSquares.svelte'
    import FileDiffHunks from './FileDiffHunks.svelte'

    export let fileDiff: FileDiffFields
    export let expanded = !!fileDiff.newPath

    $: isBinary = fileDiff.newFile?.binary
    $: isNew = !fileDiff.oldPath
    $: isDeleted = !fileDiff.newPath
    $: isRenamed = fileDiff.newPath && fileDiff.oldPath && fileDiff.newPath !== fileDiff.oldPath
    $: path = isRenamed ? `${fileDiff.oldPath} -> ${fileDiff.newPath}` : isDeleted ? fileDiff.oldPath : fileDiff.newPath
    $: badgeLabel = isNew
        ? 'Added'
        : isDeleted
        ? 'Deleted'
        : isRenamed
        ? dirname(fileDiff.newPath!) !== dirname(fileDiff.oldPath!)
            ? 'Moved'
            : 'Renamed'
        : ''
    $: stat = fileDiff.stat
    $: linkFile = fileDiff.mostRelevantFile.__typename === 'GitBlob'
</script>

<div class="header">
    <button type="button" on:click={() => (expanded = !expanded)}>
        <Icon inline svgPath={expanded ? mdiChevronDown : mdiChevronRight} />
    </button>
    <div class="headerPathStart">
        <span>{badgeLabel}</span>
    </div>
    {#if stat}
        <small>{numberWithCommas(stat.added + stat.deleted)}</small>
        <DiffSquares added={stat.added} deleted={stat.deleted} />
    {/if}
    {#if linkFile}
        <a class="file-link" href={fileDiff.mostRelevantFile.url}><strong><span title={path}>{path}</span></strong></a>
    {:else}
        <span title={path}>{path}</span>
    {/if}
</div>
{#if !isBinary && expanded}
    <FileDiffHunks hunks={fileDiff.hunks} />
{/if}

<style lang="scss">
    .header {
        display: flex;
        align-items: center;
        padding: 0.25rem 0rem;
    }

    .file-link {
        margin-left: 0.5rem;
    }

    button {
        background-color: transparent;
        border: none;
        cursor: pointer;
    }
</style>
