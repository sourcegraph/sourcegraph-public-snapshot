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
    export let expanded: boolean | undefined = undefined

    const dispatch = createEventDispatcher<{ toggle: { expanded: boolean } }>()

    $: ({ oldFile, newFile, stat } = fileDiff)

    $: isBinary = (newFile && newFile.binary) || !!oldFile?.binary
    $: isNew = !oldFile
    $: isDeleted = !newFile
    $: isRenamed = oldFile?.path !== newFile?.path
    $: isMoved = isRenamed && oldFile && newFile && dirname(newFile.path) !== dirname(oldFile.path)
    $: linkFile = fileDiff.mostRelevantFile.__typename === 'GitBlob'
    $: open = expanded === undefined ? !!fileDiff.newFile?.path && fileDiff.hunks.length > 0 : expanded

    function toggle() {
        dispatch('toggle', { expanded: !open })
    }
</script>

<div class="header">
    <Button variant="icon" on:click={toggle} aria-label="{open ? 'Hide' : 'Show'} file diff">
        <Icon inline icon={open ? ILucideChevronDown : ILucideChevronRight} />
    </Button>
    {#if isNew}
        <Badge variant="success">Added</Badge>
    {:else if isDeleted}
        <Badge variant="danger">Deleted</Badge>
    {:else if isRenamed}
        <Badge variant="warning">{isMoved ? 'Moved' : 'Renamed'}</Badge>
    {/if}
    {#if stat}
        <small class="added">+{numberWithCommas(stat.added)}<span class="visually-hidden">lines added</span></small>
        <small class="deleted"
            >-{numberWithCommas(stat.deleted)}<span class="visually-hidden">lines removed</span></small
        >
        <DiffSquares added={stat.added} deleted={stat.deleted} />
    {/if}
    {#if linkFile}
        {#if oldFile && newFile && isRenamed}
            <span>
                <a href={oldFile.canonicalURL}>{oldFile.path}</a>
                <Icon icon={ILucideArrowRight} inline aria-hidden />
                <span class="visually-hidden">{isMoved ? 'moved' : 'renamed'} to</span>
                <a href={newFile.canonicalURL}>{newFile.path}</a>
            </span>
        {:else}
            <a href={fileDiff.mostRelevantFile.url}>{fileDiff.mostRelevantFile.path}</a>
        {/if}
    {:else}
        {newFile?.path || oldFile?.path}
    {/if}
</div>
{#if open}
    {#if isBinary}
        <small class="info">(binary file not rendered)</small>
    {:else if fileDiff.hunks.length === 0}
        <small>(no changes)</small>
    {:else}
        <div class="hunks">
            <FileDiffHunks hunks={fileDiff.hunks} />
        </div>
    {/if}
{/if}

<style lang="scss">
    .header {
        position: sticky;
        top: 0;
        background-color: var(--body-bg);
        padding: 0.25rem 0rem;

        display: flex;
        align-items: center;
        gap: 0.5rem;
    }

    .hunks {
        border: 1px solid var(--border-color);
        border-radius: var(--border-radius);
        // This prevents the inner element from leaking over the rounded border
        // but also ensures that hunks are horizontally scrollable.
        overflow: auto hidden;
    }

    .added {
        color: var(--success);
    }

    .deleted {
        color: var(--danger);
    }

    small {
        color: var(--text-muted);
    }
</style>
