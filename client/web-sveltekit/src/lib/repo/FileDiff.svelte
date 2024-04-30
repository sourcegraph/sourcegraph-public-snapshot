<script lang="ts">
    import { dirname } from 'path'

    import { mdiChevronRight, mdiChevronDown } from '@mdi/js'

    import { numberWithCommas } from '$lib/common'
    import Icon from '$lib/Icon.svelte'

    import DiffSquares from './DiffSquares.svelte'
    import FileDiffHunks from './FileDiffHunks.svelte'
    import type { FileDiff_Diff } from './FileDiff.gql'
    import { createEventDispatcher } from 'svelte'

    export let fileDiff: FileDiff_Diff
    export let expanded = !!fileDiff.newPath

    const dispatch = createEventDispatcher<{ toggle: { expanded: boolean } }>()

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

    function toggle() {
        expanded = !expanded
        dispatch('toggle', { expanded })
    }
</script>

<div class="header">
    <button type="button" on:click={toggle}>
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
    <div class="hunks">
        <FileDiffHunks hunks={fileDiff.hunks} />
    </div>
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

    .hunks {
        border-radius: var(--border-radius);
        border: 1px solid var(--border-color);
    }

    button {
        background-color: transparent;
        border: none;
        cursor: pointer;
    }
</style>
