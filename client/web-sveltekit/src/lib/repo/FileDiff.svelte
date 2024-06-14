<script lang="ts">
    import { dirname } from 'path'

    import { createEventDispatcher } from 'svelte'

    import { numberWithCommas } from '$lib/common'
    import Icon from '$lib/Icon.svelte'
    import Badge from '$lib/wildcard/Badge.svelte'
    import Button from '$lib/wildcard/Button.svelte'

    import DiffSquares from './DiffSquares.svelte'
    import type { FileDiff_Diff } from './FileDiff.gql'
    import FileDiffHunks from './FileDiffHunks.svelte'

    export let fileDiff: FileDiff_Diff
    export let expanded = !!fileDiff.newPath

    const dispatch = createEventDispatcher<{ toggle: { expanded: boolean } }>()

    $: isBinary = fileDiff.newFile?.binary
    $: isNew = !fileDiff.oldPath
    $: isDeleted = !fileDiff.newPath
    $: isRenamed = fileDiff.newPath && fileDiff.oldPath && fileDiff.newPath !== fileDiff.oldPath
    $: isMoved = isRenamed && dirname(fileDiff.newPath!) !== dirname(fileDiff.oldPath!)
    $: path = isRenamed ? `${fileDiff.oldPath} -> ${fileDiff.newPath}` : isDeleted ? fileDiff.oldPath : fileDiff.newPath
    $: stat = fileDiff.stat
    $: linkFile = fileDiff.mostRelevantFile.__typename === 'GitBlob'

    function toggle() {
        expanded = !expanded
        dispatch('toggle', { expanded })
    }
</script>

<div class="header">
    <Button variant="icon" on:click={toggle} aria-label="{expanded ? 'Hide' : 'Show'} file diff">
        <Icon inline icon={expanded ? ILucideChevronDown : ILucideChevronRight} />
    </Button>
    {#if isNew}
        <Badge variant="success">Added</Badge>
    {:else if isDeleted}
        <Badge variant="danger">Deleted</Badge>
    {:else if isRenamed}
        <Badge variant="warning">{isMoved ? 'Moved' : 'Renamed'}</Badge>
    {/if}
    {#if stat}
        <small class="added">+{numberWithCommas(stat.added)}</small>
        <small class="deleted">-{numberWithCommas(stat.deleted)}</small>
        <DiffSquares added={stat.added} deleted={stat.deleted} />
    {/if}
    {#if linkFile}
        <a href={fileDiff.mostRelevantFile.url}><span title={path}>{path}</span></a>
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
        gap: 0.5rem;
        padding: 0.25rem 0rem;
    }

    .hunks {
        border-radius: var(--border-radius);
        border: 1px solid var(--border-color);
    }

    .added {
        color: var(--success);
    }

    .deleted {
        color: var(--danger);
    }
</style>
